package genrpc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
)

func TestGenServerFormatsRegisterLinesForEachScope(t *testing.T) {
	origConfig := config.C
	origTemplateHome := embeded.Home
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	templateHome, err := filepath.Abs(filepath.Join("..", "..", "..", "..", ".template"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		config.C = origConfig
		embeded.Home = origTemplateHome
		_ = os.Chdir(origWd)
	})

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "internal", "server"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	config.C = config.Config{}
	embeded.Home = templateHome

	jr := PzeroRpc{Module: "example.com/service"}
	registerLine := "device.RegisterDeviceServiceServer(grpcServer, devicesvr.NewDeviceService(ctx))"
	if err := jr.genServer(
		ImportLines{`devicesvr "example.com/service/internal/server/device"`},
		ImportLines{`device "example.com/contracts/device"`},
		RegisterLines{registerLine},
	); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(tmpDir, "internal", "server", "server.go"))
	if err != nil {
		t.Fatal(err)
	}
	contents := string(got)
	if !strings.Contains(contents, "\t\t"+registerLine+"\n") {
		t.Fatalf("RegisterZrpc callback registration is not formatted with two tabs:\n%s", contents)
	}
	if !strings.Contains(contents, "\t"+registerLine+"\n}") {
		t.Fatalf("RegisterZrpcServer registration is not formatted with one tab:\n%s", contents)
	}
	if strings.Contains(contents, "func RegisterZrpcServer(grpcServer *grpc.Server, ctx *svc.ServiceContext) {\n\t\t"+registerLine) {
		t.Fatalf("RegisterZrpcServer retained generator indentation noise:\n%s", contents)
	}
}

func TestRewriteGeneratedPBImport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "device_service_server.go")
	const goPackage = "github.com/example/contracts/gen/device/v1/device"
	const leakedImport = "github.com/example/service/tmp/pzero-rpc-123/pbout/" + goPackage
	contents := "package server\n\nimport v1_devicev1 \"" + leakedImport + "\"\n\nvar _ v1_devicev1.Device\n"
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := rewriteGeneratedPBImport(path, goPackage, "devicev1", "v1_devicev1"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "/pbout/") {
		t.Fatalf("temporary pbout import leaked into server:\n%s", got)
	}
	if !strings.Contains(string(got), `devicev1 "`+goPackage+`"`) {
		t.Fatalf("canonical external PB import missing:\n%s", got)
	}
	if !strings.Contains(string(got), "var _ devicev1.Device") {
		t.Fatalf("generated protobuf qualifier was not normalized:\n%s", got)
	}
}
