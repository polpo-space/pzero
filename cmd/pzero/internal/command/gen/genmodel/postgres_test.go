package genmodel

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Run against a disposable PostgreSQL database, using the pinned goctl binary.
func TestPostgresGeneratedModels(t *testing.T) {
	dsn, goctl := os.Getenv("PZERO_TEST_POSTGRES_DSN"), os.Getenv("GOCTL")
	if dsn == "" || goctl == "" {
		t.Skip("set PZERO_TEST_POSTGRES_DSN and GOCTL for the PostgreSQL generator contract")
	}
	root, err := filepath.Abs("../../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	schemaSQL, err := os.ReadFile("testdata/aftersale.sql")
	if err != nil {
		t.Fatal(err)
	}
	tests, err := os.ReadFile("testdata/postgres_behavior_test.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if err := os.Symlink(goctl, filepath.Join(bin, "goctl")); err != nil {
		t.Fatal(err)
	}
	pzero := filepath.Join(bin, "pzero")
	build := exec.CommandContext(ctx, "go", "build", "-o", pzero, "./cmd/pzero")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build pzero: %v\n%s", err, output)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GOCTL_EXPERIMENTAL", "on")

	for _, variant := range []struct {
		name, driver   string
		cached, mapped bool
	}{
		{name: "uncached"}, {name: "cached", driver: "postgres", cached: true}, {name: "mapped", driver: "pgx", mapped: true},
	} {
		t.Run(variant.name, func(t *testing.T) {
			schema := fmt.Sprintf("pzero_%s_%d", variant.name, time.Now().UnixNano())
			if _, err := db.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if _, err := db.ExecContext(cleanup, `DROP SCHEMA "`+schema+`" CASCADE`); err != nil {
					t.Errorf("drop test schema: %v", err)
				}
			})
			parsed, err := url.Parse(dsn)
			if err != nil {
				t.Fatal(err)
			}
			query := parsed.Query()
			query.Set("search_path", schema)
			parsed.RawQuery = query.Encode()
			scopedDSN := parsed.String()
			scoped, err := sql.Open("pgx", scopedDSN)
			if err != nil {
				t.Fatal(err)
			}
			defer scoped.Close()
			if _, err := scoped.ExecContext(ctx, string(schemaSQL)); err != nil {
				t.Fatal(err)
			}

			project, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			write := func(name, content string) {
				t.Helper()
				path := filepath.Join(project, name)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			write("go.mod", "module example.com/modeltest\n\ngo 1.25.0\n\nrequire github.com/lib/pq v1.10.9\n")
			workspace := filepath.Join(project, "go.work")
			workspaceContent := fmt.Sprintf("go 1.25.0\n\nuse (\n%q\n%q\n)\n", project, root)
			write("go.work", workspaceContent)
			write("model_test.go", string(tests))
			if variant.mapped {
				write("goctl.yaml", "model:\n  types_map:\n    _text:\n      type: types.StringList\n      null_type: types.StringList\n      pkg: example.com/modeltest/types\n")
				write("types/strings.go", `package types
import "github.com/lib/pq"
type StringList = pq.StringArray
`)
			}
			// An existing empty home keeps user-level templates out of this contract.
			if err := os.Mkdir(filepath.Join(project, ".template"), 0o755); err != nil {
				t.Fatal(err)
			}
			write(".pzero.yaml", fmt.Sprintf(`style: go_zero
gen:
  model-driver: pgx
  model-datasource: true
  model-schema: %s
  model-datasource-url: [%q]
  model-datasource-table: [tickets, ticket_messages, typed_arrays]
  model-ignore-columns: [created_at, updated_at]
  model-cache: %t
  model-cache-table: ["*"]
  model-cache-prefix: modeltest
`, schema, scopedDSN, variant.cached))
			generate := func() {
				t.Helper()
				cmd := exec.CommandContext(ctx, pzero, "gen", "model")
				cmd.Dir = project
				cmd.Env = append(cmd.Environ(), "GOWORK="+workspace)
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("pzero gen model: %v\n%s", err, output)
				}
			}
			generate()
			// An existing custom method may have the same name as a new generated
			// method. The generated implementation must live on the embedded model.
			customPath := filepath.Join(project, "internal/model/tickets/tickets_model.go")
			custom, err := os.ReadFile(customPath)
			if err != nil {
				t.Fatal(err)
			}
			custom = bytes.Replace(custom, []byte("import ("), []byte("import (\n\"context\"\n\"github.com/polpo-space/pzero/core/stores/condition\""), 1)
			custom = append(custom, []byte(`
func (m *customTicketsModel) UpdateFieldsByConditionAffected(ctx context.Context, session sqlx.Session, data map[string]any, conditions ...condition.Condition) (int64, error) {
    return m.defaultTicketsModel.UpdateFieldsByConditionAffected(ctx, session, data, conditions...)
}
`)...)
			write("internal/model/tickets/tickets_model.go", string(custom))
			generatedPath := filepath.Join(project, "internal/model/tickets/tickets_model_gen.go")
			before, err := os.ReadFile(generatedPath)
			if err != nil {
				t.Fatal(err)
			}
			generate()
			preserved, err := os.ReadFile(customPath)
			if err != nil || !bytes.Equal(custom, preserved) {
				t.Fatalf("regeneration overwrote custom model: %v", err)
			}
			after, err := os.ReadFile(generatedPath)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("regeneration is not stable: %v", err)
			}
			if variant.mapped && !strings.Contains(string(after), "types.StringList") {
				t.Fatal("goctl.yaml type override was ignored")
			}
			cmd := exec.CommandContext(ctx, "go", "test", "-race", "-count=1", "./...")
			cmd.Dir = project
			cmd.Env = append(cmd.Environ(), "GOWORK="+workspace, "PZERO_TEST_POSTGRES_DSN="+scopedDSN, "PZERO_TEST_DRIVER="+variant.driver)
			workspaceAfter, err := os.ReadFile(workspace)
			if err != nil {
				t.Fatalf("read go.work after generation: %v", err)
			}
			env := exec.CommandContext(ctx, "go", "env", "GOWORK", "GOMOD", "GOVERSION")
			env.Dir, env.Env = cmd.Dir, cmd.Env
			envOutput, err := env.CombinedOutput()
			if err != nil {
				t.Fatalf("inspect generated workspace: %v\n%s", err, envOutput)
			}
			t.Logf("go env GOWORK GOMOD GOVERSION:\n%s\ngo.work:\n%s", envOutput, workspaceAfter)
			module := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Dir}}", "github.com/polpo-space/pzero")
			module.Dir, module.Env = cmd.Dir, cmd.Env
			moduleOutput, err := module.CombinedOutput()
			if err != nil || strings.TrimSpace(string(moduleOutput)) != root {
				t.Fatalf("workspace does not resolve checkout %s: %v\n%s", root, err, moduleOutput)
			}
			if string(workspaceAfter) != workspaceContent {
				t.Fatalf("generator changed go.work:\n%s", workspaceAfter)
			}
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("generated PostgreSQL model contract: %v\n%s", err, output)
			}
		})
	}
}
