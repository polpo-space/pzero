package genapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

func TestNormalizeAPITrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.api")
	if err := os.WriteFile(path, []byte("syntax = \"v1\"\n\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := normalizeAPITrailingNewline(path); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "syntax = \"v1\"\n"; string(got) != want {
		t.Fatalf("normalized API = %q, want %q", got, want)
	}
}

func TestCollectRoutesGoBodyResolvesWorkspaceOnce(t *testing.T) {
	original := routePackageResolver
	t.Cleanup(func() { routePackageResolver = original })

	calls := 0
	routePackageResolver = func(workDir, module string) (string, string, error) {
		calls++
		return "example.com/app", "example.com/app", nil
	}

	ja := &PzeroApi{Module: "example.com/app"}
	files := []string{"desc/api/one.api", "desc/api/two.api"}
	apiSpecMap := map[string]*spec.ApiSpec{
		files[0]: {},
		files[1]: {},
	}
	currentRoutesMap := map[string][]spec.Route{
		files[0]: nil,
		files[1]: nil,
	}

	got, err := ja.collectRoutesGoBody(files, apiSpecMap, currentRoutesMap, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("routes body = %q, want empty", got)
	}
	if calls != 1 {
		t.Fatalf("workspace package resolver called %d times, want 1", calls)
	}
}
