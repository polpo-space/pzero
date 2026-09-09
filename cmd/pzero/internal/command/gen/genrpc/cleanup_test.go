package genrpc

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
)

func TestCleanupGoctlFrameArtifactsRemovesPackageNamedConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	oldStyle := config.C.Style
	config.C.Style = "go_zero"
	t.Cleanup(func() { config.C.Style = oldStyle })

	if err := os.MkdirAll(filepath.Join("cmd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join("etc"), 0o755); err != nil {
		t.Fatal(err)
	}

	generated := []string{
		"device.go",
		filepath.Join("etc", "device.yaml"),
		filepath.Join("etc", "device.v1.yaml"),
	}
	for _, path := range generated {
		if err := os.WriteFile(path, []byte("generated\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	keep := filepath.Join("etc", "etc.yaml")
	if err := os.WriteFile(keep, []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	jr := &PzeroRpc{}
	jr.cleanupGoctlFrameArtifacts("../../../contracts/proto/device/v1/device.proto", "device.v1")

	for _, path := range generated {
		if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("generated artifact %s still exists: %v", path, err)
		}
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("canonical config should remain: %v", err)
	}
}
