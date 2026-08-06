package genrpc

import "testing"

func TestIsExternalGoPackage(t *testing.T) {
	if isExternalGoPackage("./types/version") {
		t.Fatal("relative go_package should be local")
	}
	if isExternalGoPackage("") {
		t.Fatal("empty go_package should be local")
	}
	if !isExternalGoPackage("github.com/example/contracts/gen/user/v1") {
		t.Fatal("absolute go_package should be external")
	}
}

func TestResolveGoPackageImport(t *testing.T) {
	module := "github.com/example/svc"

	got := resolveGoPackageImport(module, "./types/version")
	want := "github.com/example/svc/internal/types/version"
	if got != want {
		t.Fatalf("local relative: got %s want %s", got, want)
	}

	external := "github.com/example/contracts/gen/user/v1"
	got = resolveGoPackageImport(module, external)
	if got != external {
		t.Fatalf("external absolute: got %s want %s", got, external)
	}
}
