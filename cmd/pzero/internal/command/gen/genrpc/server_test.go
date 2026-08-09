package genrpc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteGeneratedPBImport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "device_service_server.go")
	const goPackage = "github.com/example/contracts/gen/device/v1/device"
	const leakedImport = "github.com/example/service/tmp/pzero-rpc-123/pbout/" + goPackage
	contents := "package server\n\nimport device \"" + leakedImport + "\"\n"
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := rewriteGeneratedPBImport(path, goPackage); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "/pbout/") {
		t.Fatalf("temporary pbout import leaked into server:\n%s", got)
	}
	if !strings.Contains(string(got), `device "`+goPackage+`"`) {
		t.Fatalf("canonical external PB import missing:\n%s", got)
	}
}
