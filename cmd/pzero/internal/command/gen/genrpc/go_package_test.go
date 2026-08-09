package genrpc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/desc/protoparse"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
)

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
	if !isExternalGoPackage("github.com/example/contracts/gen/user/v1;userv1") {
		t.Fatal("absolute go_package with package alias should be external")
	}
}

func TestProtoImportNamePrefersBufStylePath(t *testing.T) {
	dirs := []string{"../../../contracts/proto/user"}
	got, err := protoImportName("../../../contracts/proto/user/v1/user_messages.proto", dirs)
	if err != nil {
		t.Fatal(err)
	}
	if got != "user/v1/user_messages.proto" {
		t.Fatalf("got %q want user/v1/user_messages.proto", got)
	}
}

func TestRelToProtoDirUsesCanonicalImportName(t *testing.T) {
	root := t.TempDir()
	protoDir := filepath.Join(root, "device", "v1")
	if err := os.MkdirAll(protoDir, 0o755); err != nil {
		t.Fatal(err)
	}

	deviceFile := filepath.Join(protoDir, "device.proto")
	regionFile := filepath.Join(protoDir, "region.proto")
	if err := os.WriteFile(deviceFile, []byte(`syntax = "proto3";
package device.v1;
import "device/v1/region.proto";
message DeviceRequest { region.v1.Region region = 1; }
message DeviceResponse {}
service DeviceService { rpc Get(DeviceRequest) returns (DeviceResponse); }
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(regionFile, []byte(`syntax = "proto3";
package region.v1;
message Region { string code = 1; }
message RegionRequest {}
message RegionResponse {}
service RegionService { rpc Get(RegionRequest) returns (RegionResponse); }
`), 0o644); err != nil {
		t.Fatal(err)
	}

	previousIncludes := config.C.Gen.ProtoInclude
	config.C.Gen.ProtoInclude = []string{root}
	t.Cleanup(func() { config.C.Gen.ProtoInclude = previousIncludes })

	protoDirs := []string{protoDir}
	parser := protoparse.Parser{
		ImportPaths:           buildProtoImportPaths(protoDirs),
		InferImportPaths:      false,
		IncludeSourceCodeInfo: true,
	}

	var names []string
	for _, file := range []string{deviceFile, regionFile} {
		_, name, err := relToProtoDir(file, protoDirs)
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	wantNames := []string{"device/v1/device.proto", "device/v1/region.proto"}
	for i := range wantNames {
		if names[i] != wantNames[i] {
			t.Fatalf("proto name %d: got %q want %q", i, names[i], wantNames[i])
		}
	}
	if _, err := parser.ParseFiles(names...); err != nil {
		t.Fatalf("parse service protos as %q: %v", names, err)
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

	aliased := external + ";userv1"
	got = resolveGoPackageImport(module, aliased)
	if got != external {
		t.Fatalf("aliased external import: got %s want %s", got, external)
	}
	if got := resolveGoPackageMapping(module, aliased); got != aliased {
		t.Fatalf("aliased external mapping: got %s want %s", got, aliased)
	}
	if got := goPackageName(aliased, "v1"); got != "userv1" {
		t.Fatalf("aliased package name: got %s want userv1", got)
	}
}
