package genmodel

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"text/template"

	goctlconfig "github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/gen"
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/model"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

func TestPostgresModelTemplates(t *testing.T) {
	home, err := filepath.Abs("../../../../.template/go-zero")
	if err != nil {
		t.Fatal(err)
	}
	oldHome, err := pathx.GetGoctlHome()
	if err != nil {
		t.Fatal(err)
	}
	pathx.RegisterGoctlHome(home)
	t.Cleanup(func() { pathx.RegisterGoctlHome(oldHome) })

	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "go.mod"), []byte("module example.com/modeltest\n\ngo 1.25.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Compile against this checkout without changing its go.work or module files.
	workspace := filepath.Join(project, "go.work")
	work := fmt.Sprintf("go 1.25.0\n\nuse (\n%q\n%q\n)\n", project, filepath.Clean(filepath.Join(home, "../../../..")))
	if err := os.WriteFile(workspace, []byte(work), 0o600); err != nil {
		t.Fatal(err)
	}

	behavior, err := template.ParseFiles("testdata/postgres_model_test.go.tpl")
	if err != nil {
		t.Fatal(err)
	}

	for _, autoIncrement := range []bool{false, true} {
		for _, cached := range []bool{false, true} {
			for _, nullable := range []bool{false, true} {
				t.Run(fmt.Sprintf("auto=%t/cache=%t/nullable=%t", autoIncrement, cached, nullable), func(t *testing.T) {
					primary := &model.Column{DbColumn: &model.DbColumn{
						Name: "device_id", DataType: "bigint", IsNullAble: "NO", OrdinalPosition: 1,
					}}
					if autoIncrement {
						primary.Extra = "auto_increment"
					}
					state := &model.Column{DbColumn: &model.DbColumn{
						Name: "state", DataType: "int", IsNullAble: "NO", OrdinalPosition: 2,
					}}
					if nullable {
						state.IsNullAble = "YES"
					}
					table := &model.Table{
						Db: "public", Table: "device_states", PrimaryKey: primary,
						Columns:     []*model.Column{primary, state},
						UniqueIndex: map[string][]*model.Column{"state": {state}},
					}
					dir := filepath.Join(project, fmt.Sprintf("auto_%t_cache_%t_nullable_%t", autoIncrement, cached, nullable))
					generator, err := gen.NewDefaultGenerator("cache", dir, &goctlconfig.Config{NamingFormat: "go_zero"}, gen.WithPostgreSql())
					if err != nil {
						t.Fatal(err)
					}
					if err := generator.StartFromInformationSchema(map[string]*model.Table{table.Table: table}, cached, true); err != nil {
						t.Fatal(err)
					}
					if !nullable {
						var test bytes.Buffer
						if err := behavior.Execute(&test, map[string]any{
							"Package": filepath.Base(dir), "Cached": cached, "AutoIncrement": autoIncrement,
						}); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(filepath.Join(dir, "postgres_test.go"), test.Bytes(), 0o600); err != nil {
							t.Fatal(err)
						}
					}
					filename := filepath.Join(dir, "device_states_model_gen.go")
					file, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.ImportsOnly)
					if err != nil {
						t.Fatal(err)
					}
					var importsSQL bool
					for _, imp := range file.Imports {
						importsSQL = importsSQL || imp.Path.Value == `"database/sql"`
					}
					if want := cached || nullable; importsSQL != want {
						t.Fatalf("database/sql imported = %t, want %t", importsSQL, want)
					}
				})
			}
		}
	}
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = project
	cmd.Env = append(cmd.Environ(), "GOWORK="+workspace)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated model checks failed: %v\n%s", err, output)
	}
}
