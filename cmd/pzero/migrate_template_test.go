package main

import (
	"strings"
	"testing"

	pzerotemplatex "github.com/polpo-space/pzero/cmd/pzero/internal/pkg/templatex"
)

func TestAPIRPCTemplatesProvideServiceMigrateCommand(t *testing.T) {
	for _, frame := range []string{"api", "rpc"} {
		t.Run(frame, func(t *testing.T) {
			commandPath := ".template/frame/" + frame + "/app/cmd/migrate.go.tpl"
			commandTemplate := readEmbeddedTemplate(t, commandPath)
			for _, want := range []string{
				`{{ if has "model" .Features }}`,
				`github.com/polpo-space/pzero/core/migrator`,
				`migrator.NewCommand`,
				`conf.Load(configPath, &c, conf.UseEnv())`,
				`return c.Sqlx.SqlConf, nil`,
			} {
				if !strings.Contains(commandTemplate, want) {
					t.Fatalf("%s migrate command template is missing %q", frame, want)
				}
			}
			assertTemplateEmptyWithoutModel(t, commandPath, commandTemplate)

			serverTemplate := readEmbeddedTemplate(t,
				".template/frame/"+frame+"/app/cmd/server.go.tpl")
			if strings.Contains(serverTemplate, "migrator.New") ||
				strings.Contains(serverTemplate, "m.Up()") {
				t.Fatalf("%s server template still runs migrations during startup", frame)
			}

			for _, direction := range []string{"up", "down"} {
				migrationPath := ".template/frame/" + frame + "/app/desc/sql_migration/1_initialize_schema." + direction + ".sql.tpl"
				migrationTemplate := readEmbeddedTemplate(t, migrationPath)
				if !strings.Contains(migrationTemplate, `if has "model" .Features`) {
					t.Fatalf("%s %s migration template is not model-gated", frame, direction)
				}
				if strings.Contains(strings.ToLower(migrationTemplate), "select 1") {
					t.Fatalf("%s %s migration template executes placeholder SQL", frame, direction)
				}
				if !strings.Contains(migrationTemplate, "-- Write your "+direction+" migration SQL here.") {
					t.Fatalf("%s %s migration template is missing its placeholder comment", frame, direction)
				}
				assertTemplateEmptyWithoutModel(t, migrationPath, migrationTemplate)

				rendered, err := pzerotemplatex.ParseTemplate(migrationPath, map[string]any{
					"Features": []string{"model"},
				}, []byte(migrationTemplate))
				if err != nil {
					t.Fatalf("render %s %s migration template: %v", frame, direction, err)
				}
				want := "-- Write your " + direction + " migration SQL here.\n"
				if string(rendered) != want {
					t.Fatalf("%s %s migration template rendered %q, want %q", frame, direction, rendered, want)
				}
			}

			readmeTemplate := readEmbeddedTemplate(t,
				".template/frame/"+frame+"/app/README.md.tpl")
			if !strings.Contains(readmeTemplate, "go run . migrate up --config etc/etc.yaml") {
				t.Fatalf("%s README template does not document explicit service migrations", frame)
			}
		})
	}
}

func assertTemplateEmptyWithoutModel(t *testing.T, path, content string) {
	t.Helper()

	rendered, err := pzerotemplatex.ParseTemplate(path, map[string]any{
		"Features": []string{},
	}, []byte(content))
	if err != nil {
		t.Fatalf("render template %s without model: %v", path, err)
	}
	if len(rendered) != 0 {
		t.Fatalf("template %s rendered %d bytes without model: %q", path, len(rendered), rendered)
	}
}
