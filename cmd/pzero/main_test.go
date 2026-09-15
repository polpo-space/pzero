package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	genrun "github.com/polpo-space/pzero/cmd/pzero/internal/command/gen/gen"
	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
)

func TestGeneratorRegeneratesSharedAPITypes(t *testing.T) {
	goctl := os.Getenv("GOCTL")
	if goctl == "" {
		t.Skip("set GOCTL to run the real API generator regression")
	}
	t.Setenv("PATH", filepath.Dir(goctl)+string(os.PathListSeparator)+os.Getenv("PATH"))
	// The fixture has no workspace; clear any inherited workspace path.
	t.Setenv("GOWORK", "")
	oldConfig, oldTemplates, oldHome := config.C, embeded.Template, embeded.Home
	t.Cleanup(func() { config.C, embeded.Template, embeded.Home = oldConfig, oldTemplates, oldHome })
	config.C = config.Config{Quiet: true, Style: config.DefaultStyle, Gen: config.GenConfig{ApiTypesDir: filepath.Join("internal", "types")}}
	embeded.Template, embeded.Home = Template, ""
	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(project)
	if err := os.MkdirAll("desc/api", 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(name, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/sharedapi\n\ngo 1.25.0\n")
	for _, name := range []string{"a", "b"} {
		write("desc/api/"+name+".api", "syntax = \"v1\"\nimport \"common.api\"\ninfo (go_package: \"shared\")\n@server(group: "+name+")\nservice demo {\n @handler Get\n get /"+name+" returns (Shared)\n}\n")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	metadataDir := filepath.Join(home, ".pzero", "desc-metadata", strings.TrimPrefix(filepath.ToSlash(config.C.Wd()), "/"))
	t.Cleanup(func() { _ = os.RemoveAll(metadataDir) })
	for _, field := range []string{"Original", "Updated"} {
		write("desc/api/common.api", "syntax = \"v1\"\ntype Shared {\n "+field+" string `json:\"value\"`\n}\n")
		if err := genrun.Run(); err != nil {
			t.Fatal(err)
		}
		types, err := os.ReadFile("internal/types/shared/types.go")
		if err != nil || !strings.Contains(string(types), field) || strings.Count(string(types), "type Shared struct") != 1 {
			t.Fatalf("shared types were not regenerated once: %s, %v", types, err)
		}
		if field == "Updated" && strings.Contains(string(types), "Original") {
			t.Fatal("old shared field survived regeneration")
		}
		for _, path := range []string{"internal/handler/a/get.go", "internal/handler/b/get.go"} {
			if _, err := os.Stat(path); err != nil {
				t.Fatal(err)
			}
		}
		for _, path := range []string{metadataDir, "internal/handler/route2code.go"} {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("retired artifact %s exists or cannot be checked: %v", path, err)
			}
		}
	}
}
