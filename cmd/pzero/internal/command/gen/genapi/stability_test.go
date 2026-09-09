package genapi

import (
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apiparser "github.com/zeromicro/go-zero/tools/goctl/api/parser"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
)

func TestPackageTypesAndConflictPreflight(t *testing.T) {
	setTypesDir(t, defaultTypesDir)
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("desc/api", 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join("desc/api", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("common.api", "syntax = \"v1\"\ntype Shared {\n Value string `json:\"value\"`\n}\n")
	files := []string{"desc/api/a.api", "desc/api/b.api"}
	specs := make(map[string]*spec.ApiSpec)
	for _, name := range []string{"a", "b"} {
		write(name+".api", "syntax = \"v1\"\nimport \"common.api\"\ninfo (go_package: \"shared\")\n@server(group: "+name+")\nservice demo {\n @handler Get\n get /"+name+" returns (Shared)\n}\n")
		parsed, err := apiparser.Parse("desc/api/" + name + ".api")
		if err != nil {
			t.Fatal(err)
		}
		specs["desc/api/"+name+".api"] = parsed
	}
	ja := &PzeroApi{}
	if err := validateAPITypes(files, specs); err != nil {
		t.Fatal(err)
	}
	if err := ja.separateTypesGo(files, specs); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile("internal/types/shared/types.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(output), "type Shared struct") != 1 {
		t.Fatalf("duplicate declarations: %s", output)
	}
	// Conflicting imported definitions must fail before touching existing output.
	write("b.api", "syntax = \"v1\"\ninfo (go_package: \"shared\")\ntype Shared {\n Value int64 `json:\"value\"`\n}\n@server(group: b)\nservice demo {\n @handler Get\n get /b returns (Shared)\n}\n")
	snapshot := func() map[string]string {
		t.Helper()
		files := make(map[string]string)
		if err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			contents, err := os.ReadFile(path)
			files[path] = string(contents)
			return err
		}); err != nil {
			t.Fatal(err)
		}
		return files
	}
	before := snapshot()
	if _, err := ja.Gen(nil); err == nil || !strings.Contains(err.Error(), "conflicting API type") {
		t.Fatalf("expected type conflict, got %v", err)
	}
	if !maps.Equal(before, snapshot()) {
		t.Fatal("conflict modified project files")
	}
}

func TestUniqueAPITypesIgnoresDocumentationButChecksTags(t *testing.T) {
	a := spec.DefineStruct{RawName: "Item", Members: []spec.Member{{Name: "Value", Type: spec.PrimitiveType{RawName: "string"}, Tag: "`json:\"value\"`"}}}
	a.Members = append(a.Members, spec.Member{Name: "Other", Type: spec.PrimitiveType{RawName: "string"}})
	b := a
	b.Docs = spec.Doc{"// Different documentation"}
	b.Members = append([]spec.Member(nil), a.Members...)
	b.Members[1].Docs = spec.Doc{"// Several lines", "// of field documentation"}
	got, err := uniqueAPITypes([]spec.Type{a, b})
	if err != nil || len(got) != 1 {
		t.Fatalf("dedup: %v %v", got, err)
	}
	b.Members = append([]spec.Member(nil), a.Members...)
	b.Members[0].Tag = "`json:\"other\"`"
	if _, err := uniqueAPITypes([]spec.Type{a, b}); err == nil {
		t.Fatal("conflicting tag accepted")
	}
	specs := map[string]*spec.ApiSpec{
		"a.api": {Info: spec.Info{Properties: map[string]string{"go_package": "a"}}, Types: []spec.Type{a}},
		"b.api": {Info: spec.Info{Properties: map[string]string{"go_package": "b"}}, Types: []spec.Type{b}},
	}
	if err := validateAPITypes([]string{"a.api", "b.api"}, specs); err != nil {
		t.Fatalf("different packages must remain independent: %v", err)
	}
	specs["a.api"].Info.Properties["go_package"] = ""
	specs["b.api"].Info.Properties["go_package"] = ""
	if err := validateAPITypes([]string{"a.api", "b.api"}, specs); err == nil {
		t.Fatal("default package conflict accepted")
	}
}

func TestWorkspaceImportResolutionDoesNotWriteGoWork(t *testing.T) {
	oldStyle := config.C.Style
	config.C.Style = "go_zero"
	t.Cleanup(func() { config.C.Style = oldStyle })
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	// A comment and layout make go work use observably different even when the
	// module is already present. Package resolution must be read-only.
	workspace := "go 1.25.0\n\nuse (\n ./service // preserve this declaration\n ./other\n)\n"
	if err := os.WriteFile("go.work", []byte(workspace), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("service", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("service/go.mod", []byte("module example.com/service\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("other", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("other/go.mod", []byte("module example.com/other\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOWORK", filepath.Join(root, "go.work"))
	t.Chdir(filepath.Join(root, "service"))
	api := "syntax = \"v1\"\nservice demo {\n @handler Get\n get /get\n}\n"
	if err := os.WriteFile("service.api", []byte(api), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed, err := apiparser.Parse("service.api")
	if err != nil {
		t.Fatal(err)
	}
	files := []string{"a.api", "b.api", "c.api", "d.api"}
	specs := make(map[string]*spec.ApiSpec)
	routes := make(map[string][]spec.Route)
	for _, file := range files {
		specs[file] = parsed
		routes[file] = parsed.Service.Routes()
	}
	ja := &PzeroApi{Module: "example.com/service"}
	if _, err := ja.collectRoutesGoBody(files, specs, routes, nil); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, "", "package handler\nimport \"example.com/service/internal/types\"\nvar _ types.Request", 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := UpdateImportedModule(node, fset, filepath.Join(root, "service"), "example.com/service"); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile(filepath.Join(root, "go.work"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != workspace {
		t.Fatalf("workspace modified: %s", got)
	}
}

func TestPatchFilesKeepsBothSharedCompactHandlers(t *testing.T) {
	t.Chdir(t.TempDir())
	setTypesDir(t, defaultTypesDir)
	oldStyle := config.C.Style
	config.C.Style = "go_zero"
	t.Cleanup(func() { config.C.Style = oldStyle })
	for _, dir := range []string{"internal/handler/shared", "internal/logic/shared"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	specs := make(map[string]*spec.ApiSpec)
	files := []string{"a.api", "b.api"}
	for _, name := range []string{"A", "B"} {
		lower := strings.ToLower(name)
		source := "syntax = \"v1\"\n@server(group: shared\n compact_handler: true)\nservice demo {\n @handler Get" + name + "\n get /" + lower + "\n}\n"
		if err := os.WriteFile(lower+".api", []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		parsed, err := apiparser.Parse(lower + ".api")
		if err != nil {
			t.Fatal(err)
		}
		specs[lower+".api"] = parsed
		handler := "package shared\nfunc Get" + name + "Handler() {}\n"
		logic := "package shared\ntype Get" + name + "Logic struct{}\nfunc (l *Get" + name + "Logic) Get" + name + "() error { return nil }\n"
		if err := os.WriteFile("internal/handler/shared/get_"+lower+"_handler.go", []byte(handler), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("internal/logic/shared/get_"+lower+"_logic.go", []byte(logic), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ja := &PzeroApi{}
	if err := ja.patchHandlerAndLogicFiles(files, specs, specs); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile("internal/handler/shared/shared_compact.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, handler := range []string{"func GetA(", "func GetB("} {
		if strings.Count(string(output), handler) != 1 {
			t.Fatalf("missing or duplicated %s: %s", handler, output)
		}
	}
}
