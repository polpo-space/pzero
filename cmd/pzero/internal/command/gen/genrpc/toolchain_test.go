package genrpc

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Run with GOCTL pointing to a binary built from this module's dependency graph.
func TestGoctlCrossProtoServices(t *testing.T) {
	goctl := os.Getenv("GOCTL")
	if goctl == "" {
		t.Skip("set GOCTL to run the real protoc compatibility regression")
	}
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOWORK", "")
	protoDir := filepath.Join(root, "proto", "device", "v1")
	if err := os.MkdirAll(protoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixtures := map[string]string{
		"region.proto": `syntax = "proto3";
package device.v1;
option go_package = "example.com/contracts/device/v1;devicev1";
message Region { string name = 1; }
service RegionService { rpc Get(Region) returns (Region); }
`,
		"device.proto": `syntax = "proto3";
package device.v1;
option go_package = "example.com/contracts/device/v1;devicev1";
import "device/v1/region.proto";
message Device { Region region = 1; }
service DeviceService { rpc Get(Device) returns (Device); }
`,
	}
	for name, body := range fixtures {
		if err := os.WriteFile(filepath.Join(protoDir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, layout := range []string{"overlapping", "canonical"} {
		t.Run(layout, func(t *testing.T) {
			out := filepath.Join(root, layout)
			if err := os.MkdirAll(filepath.Join(out, "pb"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(out, "go.mod"), []byte("module example.com/fixture\n\ngo 1.25.0\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"device.proto", "region.proto"} {
				args := []string{"rpc", "protoc", filepath.Join(protoDir, name)}
				if layout == "overlapping" {
					args = append(args, "-I", protoDir)
				}
				args = append(args, "-I", filepath.Join(root, "proto"), "--go_out=./pb", "--go-grpc_out=./pb", "--zrpc_out=.", "-m")
				cmd := exec.Command(goctl, args...)
				cmd.Dir = out
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s: %v\n%s", name, err, output)
				}
			}
		})
	}
}
