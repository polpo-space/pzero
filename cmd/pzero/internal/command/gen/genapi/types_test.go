package genapi

import (
	"bytes"
	"errors"
	goformat "go/format"
	goparser "go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apiparser "github.com/zeromicro/go-zero/tools/goctl/api/parser"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
)

func setTypesDir(t *testing.T, typesDir string) {
	t.Helper()

	oldTypesDir := config.C.Gen.ApiTypesDir
	config.C.Gen.ApiTypesDir = typesDir
	t.Cleanup(func() {
		config.C.Gen.ApiTypesDir = oldTypesDir
	})
}

func formatFile(t *testing.T, fset *token.FileSet, f any) string {
	t.Helper()

	var buf bytes.Buffer
	if err := goformat.Node(&buf, fset, f); err != nil {
		t.Fatalf("format.Node() error = %v", err)
	}

	return buf.String()
}

func TestRenderTypesFileConditionallyImportsTime(t *testing.T) {
	ja := &PzeroApi{}

	withoutTime, err := ja.renderTypesFile([]spec.Type{
		spec.DefineStruct{
			RawName: "User",
			Members: []spec.Member{
				{Name: "Name", Type: spec.PrimitiveType{RawName: "string"}},
			},
		},
	}, "types")
	if err != nil {
		t.Fatalf("render types without time: %v", err)
	}
	if strings.Contains(string(withoutTime), `"time"`) || strings.Contains(string(withoutTime), "time.Now()") {
		t.Fatalf("types without time contain a time placeholder:\n%s", withoutTime)
	}

	withTime, err := ja.renderTypesFile([]spec.Type{
		spec.DefineStruct{
			RawName: "Event",
			Members: []spec.Member{
				{Name: "CreatedAt", Type: spec.PrimitiveType{RawName: "time.Time"}},
			},
		},
	}, "types")
	if err != nil {
		t.Fatalf("render types with time: %v", err)
	}
	if !strings.Contains(string(withTime), `"time"`) || !strings.Contains(string(withTime), "CreatedAt time.Time") {
		t.Fatalf("types with time are missing the required import:\n%s", withTime)
	}
}

func TestSeparateTypesGoRewritesDefaultTypesFileWhenNoDefaultTypesRemain(t *testing.T) {
	tmpDir := t.TempDir()
	setTypesDir(t, defaultTypesDir)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	if err := os.MkdirAll(filepath.Join("internal", "types"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	oldContent := []byte(`package types

type OldType struct{}
`)
	defaultTypesPath := filepath.Join("internal", "types", "types.go")
	if err := os.WriteFile(defaultTypesPath, oldContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	apiSpecMap := map[string]*spec.ApiSpec{
		"desc/api/user.api": {
			Info: spec.Info{
				Properties: map[string]string{
					"go_package": "user",
				},
			},
			Types: []spec.Type{
				spec.DefineStruct{
					RawName: "UserReq",
					Members: []spec.Member{
						{
							Name: "Name",
							Type: spec.PrimitiveType{RawName: "string"},
							Tag:  "`json:\"name\"`",
						},
					},
				},
			},
		},
	}

	ja := &PzeroApi{}
	if err := ja.separateTypesGo([]string{"desc/api/user.api"}, apiSpecMap); err != nil {
		t.Fatalf("separateTypesGo() error = %v", err)
	}

	data, err := os.ReadFile(defaultTypesPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	got := string(data)
	if strings.Contains(got, "OldType") {
		t.Fatalf("separateTypesGo() should rewrite stale default types file, got:\n%s", got)
	}
	if !strings.Contains(got, "package types") {
		t.Fatalf("separateTypesGo() should keep default types package, got:\n%s", got)
	}
}

func TestSeparateTypesGoMergesTypesForSameGoPackage(t *testing.T) {
	tmpDir := t.TempDir()
	setTypesDir(t, defaultTypesDir)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	apiSpecMap := map[string]*spec.ApiSpec{
		"desc/api/user.api": {
			Info: spec.Info{
				Properties: map[string]string{
					"go_package": "shared",
				},
			},
			Types: []spec.Type{
				spec.DefineStruct{
					RawName: "UserReq",
					Members: []spec.Member{
						{
							Name: "Name",
							Type: spec.PrimitiveType{RawName: "string"},
							Tag:  "`json:\"name\"`",
						},
					},
				},
			},
		},
		"desc/api/order.api": {
			Info: spec.Info{
				Properties: map[string]string{
					"go_package": "shared",
				},
			},
			Types: []spec.Type{
				spec.DefineStruct{
					RawName: "OrderReq",
					Members: []spec.Member{
						{
							Name: "Id",
							Type: spec.PrimitiveType{RawName: "int64"},
							Tag:  "`json:\"id\"`",
						},
					},
				},
			},
		},
	}

	ja := &PzeroApi{}
	if err := ja.separateTypesGo([]string{"desc/api/user.api", "desc/api/order.api"}, apiSpecMap); err != nil {
		t.Fatalf("separateTypesGo() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join("internal", "types", "shared", "types.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	got := string(data)
	for _, want := range []string{
		"package shared",
		"type UserReq struct",
		"type OrderReq struct",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("separateTypesGo() missing %q in merged file:\n%s", want, got)
		}
	}
}

func TestSeparateTypesGoDeduplicatesSharedImportedTypeForSameGoPackage(t *testing.T) {
	tmpDir := t.TempDir()
	setTypesDir(t, defaultTypesDir)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	apiDir := filepath.Join("desc", "api")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	files := map[string]string{
		"common.api": `syntax = "v1"

type Shared {
    value string ` + "`json:\"value\"`" + `
}`,
		"a.api": `syntax = "v1"

import "common.api"

info (
    go_package: "shared"
)

@server (
    prefix: /a
    group: a
    compact_handler: true
)
service demo {
    @handler GetA
    get /get returns (Shared)
}`,
		"b.api": `syntax = "v1"

import "common.api"

info (
    go_package: "shared"
)

@server (
    prefix: /b
    group: b
    compact_handler: true
)
service demo {
    @handler GetB
    get /get returns (Shared)
}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(apiDir, name), []byte(content+"\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	apiSpecMap := make(map[string]*spec.ApiSpec)
	var apiFiles []string
	for _, name := range []string{"a.api", "b.api"} {
		path := filepath.Join(apiDir, name)
		parsed, err := apiparser.Parse(path)
		if err != nil {
			t.Fatalf("Parse(%s) error = %v", path, err)
		}
		apiFiles = append(apiFiles, path)
		apiSpecMap[path] = parsed
	}

	ja := &PzeroApi{}
	if err := ja.separateTypesGo(apiFiles, apiSpecMap); err != nil {
		t.Fatalf("separateTypesGo() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join("internal", "types", "shared", "types.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := strings.Count(string(data), "type Shared struct"); got != 1 {
		t.Fatalf("Shared declaration count = %d, want 1:\n%s", got, data)
	}
}

func TestSeparateTypesGoDeduplicatesIdenticalDefinitionsForSameGoPackage(t *testing.T) {
	tmpDir := t.TempDir()
	setTypesDir(t, defaultTypesDir)
	withWorkingDir(t, tmpDir)

	shared := spec.DefineStruct{
		RawName: "Shared",
		Members: []spec.Member{{
			Name: "Value",
			Type: spec.PrimitiveType{RawName: "string"},
			Tag:  "`json:\"value\"`",
		}},
	}
	apiSpecMap := map[string]*spec.ApiSpec{
		"desc/api/a.api": apiSpecWithPackage("shared", shared),
		"desc/api/b.api": apiSpecWithPackage("shared", shared),
	}

	ja := &PzeroApi{}
	if err := ja.separateTypesGo([]string{"desc/api/a.api", "desc/api/b.api"}, apiSpecMap); err != nil {
		t.Fatalf("separateTypesGo() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join("internal", "types", "shared", "types.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := strings.Count(string(data), "type Shared struct"); got != 1 {
		t.Fatalf("Shared declaration count = %d, want 1:\n%s", got, data)
	}
}

func TestSeparateTypesGoRejectsConflictingDefinitionsForSameGoPackage(t *testing.T) {
	tmpDir := t.TempDir()
	setTypesDir(t, defaultTypesDir)
	withWorkingDir(t, tmpDir)

	apiSpecMap := map[string]*spec.ApiSpec{
		"desc/api/a.api": apiSpecWithPackage("shared", spec.DefineStruct{
			RawName: "Shared",
			Members: []spec.Member{{Name: "Value", Type: spec.PrimitiveType{RawName: "string"}}},
		}),
		"desc/api/b.api": apiSpecWithPackage("shared", spec.DefineStruct{
			RawName: "Shared",
			Members: []spec.Member{{Name: "Id", Type: spec.PrimitiveType{RawName: "int64"}}},
		}),
	}

	ja := &PzeroApi{}
	err := ja.separateTypesGo([]string{"desc/api/a.api", "desc/api/b.api"}, apiSpecMap)
	if err == nil {
		t.Fatal("separateTypesGo() should reject conflicting type definitions")
	}
	for _, want := range []string{`conflicting API type "Shared"`, `go_package "shared"`, "desc/api/a.api", "desc/api/b.api"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}
}

func TestSeparateTypesGoAllowsSameTypeNameInDifferentGoPackages(t *testing.T) {
	tmpDir := t.TempDir()
	setTypesDir(t, defaultTypesDir)
	withWorkingDir(t, tmpDir)

	apiSpecMap := map[string]*spec.ApiSpec{
		"desc/api/order.api": apiSpecWithPackage("order", spec.DefineStruct{
			RawName: "Item",
			Members: []spec.Member{{Name: "Id", Type: spec.PrimitiveType{RawName: "int64"}}},
		}),
		"desc/api/device.api": apiSpecWithPackage("device", spec.DefineStruct{
			RawName: "Item",
			Members: []spec.Member{{Name: "Code", Type: spec.PrimitiveType{RawName: "string"}}},
		}),
	}

	ja := &PzeroApi{}
	if err := ja.separateTypesGo([]string{"desc/api/order.api", "desc/api/device.api"}, apiSpecMap); err != nil {
		t.Fatalf("separateTypesGo() error = %v", err)
	}
	for _, pkg := range []string{"order", "device"} {
		data, err := os.ReadFile(filepath.Join("internal", "types", pkg, "types.go"))
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", pkg, err)
		}
		if !strings.Contains(string(data), "type Item struct") {
			t.Fatalf("package %s missing Item declaration:\n%s", pkg, data)
		}
	}
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
}

func apiSpecWithPackage(goPackage string, types ...spec.Type) *spec.ApiSpec {
	return &spec.ApiSpec{
		Info:  spec.Info{Properties: map[string]string{"go_package": goPackage}},
		Types: types,
	}
}

func TestSeparateTypesGoWritesCustomTypesDirAndRemovesLegacyGeneratedTypes(t *testing.T) {
	tmpDir := t.TempDir()
	setTypesDir(t, filepath.Join("pkg", "types"))

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	legacyTypesDir := filepath.Join("internal", "types")
	if err := os.MkdirAll(legacyTypesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	legacyTypesPath := filepath.Join(legacyTypesDir, "types.go")
	if err := os.WriteFile(legacyTypesPath, []byte(`// Code generated by goctl. DO NOT EDIT.
package types

type Legacy struct{}
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	apiSpecMap := map[string]*spec.ApiSpec{
		"desc/api/user.api": {
			Types: []spec.Type{
				spec.DefineStruct{
					RawName: "UserReq",
					Members: []spec.Member{
						{
							Name: "Name",
							Type: spec.PrimitiveType{RawName: "string"},
							Tag:  "`json:\"name\"`",
						},
					},
				},
			},
		},
	}

	ja := &PzeroApi{}
	if err := ja.separateTypesGo([]string{"desc/api/user.api"}, apiSpecMap); err != nil {
		t.Fatalf("separateTypesGo() error = %v", err)
	}

	customTypesPath := filepath.Join("pkg", "types", "types.go")
	data, err := os.ReadFile(customTypesPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", customTypesPath, err)
	}

	got := string(data)
	if !strings.Contains(got, "type UserReq struct") {
		t.Fatalf("separateTypesGo() should write default types to custom dir, got:\n%s", got)
	}

	if _, err := os.Stat(legacyTypesPath); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("separateTypesGo() should remove generated legacy types file, stat err = %v", err)
	}
}

func TestUpdateHandlerImportedTypesPathUsesConfiguredTypesDir(t *testing.T) {
	setTypesDir(t, filepath.Join("pkg", "types"))

	fset := token.NewFileSet()
	src := `package user

import "example.com/demo/internal/types"

func Get() {
	var _ types.UserReq
}
`

	f, err := goparser.ParseFile(fset, "handler.go", src, goparser.ParseComments)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	ja := &PzeroApi{Module: "example.com/demo"}
	if err := ja.updateHandlerImportedTypesPath(f, fset, HandlerFile{}); err != nil {
		t.Fatalf("updateHandlerImportedTypesPath() error = %v", err)
	}

	got := formatFile(t, fset, f)
	if strings.Contains(got, `"example.com/demo/internal/types"`) {
		t.Fatalf("updateHandlerImportedTypesPath() should remove legacy import, got:\n%s", got)
	}
	if !strings.Contains(got, `types "example.com/demo/pkg/types"`) {
		t.Fatalf("updateHandlerImportedTypesPath() should use configured types dir, got:\n%s", got)
	}
}

func TestUpdateLogicImportedTypesPathUsesConfiguredTypesPackageDir(t *testing.T) {
	setTypesDir(t, filepath.Join("pkg", "types"))

	fset := token.NewFileSet()
	src := `package user

import "example.com/demo/internal/types"

func Get() {
	var _ *types.UserReq
}
`

	f, err := goparser.ParseFile(fset, "logic.go", src, goparser.ParseComments)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	ja := &PzeroApi{Module: "example.com/demo"}
	err = ja.updateLogicImportedTypesPath(f, fset, LogicFile{
		Package: "shared",
		RequestType: spec.DefineStruct{
			RawName: "UserReq",
		},
	})
	if err != nil {
		t.Fatalf("updateLogicImportedTypesPath() error = %v", err)
	}

	got := formatFile(t, fset, f)
	if strings.Contains(got, `"example.com/demo/internal/types"`) {
		t.Fatalf("updateLogicImportedTypesPath() should remove legacy import, got:\n%s", got)
	}
	if !strings.Contains(got, `types "example.com/demo/pkg/types/shared"`) {
		t.Fatalf("updateLogicImportedTypesPath() should use configured package dir, got:\n%s", got)
	}
}
