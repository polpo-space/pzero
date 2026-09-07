package config

import (
	"path/filepath"
	"testing"
)

func TestProtoDirsDefaultsToDescProto(t *testing.T) {
	c := Config{}
	dirs := c.ProtoDirs()
	if len(dirs) != 1 || dirs[0] != filepath.Join("desc", "proto") {
		t.Fatalf("unexpected default proto dirs: %v", dirs)
	}
	if c.ProtoDir() != filepath.Join("desc", "proto") {
		t.Fatalf("unexpected ProtoDir: %s", c.ProtoDir())
	}
}

func TestProtoDirsUsesConfiguredRoots(t *testing.T) {
	c := Config{
		Gen: GenConfig{
			ProtoDirs: []string{"../../contracts/proto/user", "", "."},
		},
	}
	dirs := c.ProtoDirs()
	if len(dirs) != 1 || dirs[0] != filepath.Clean("../../contracts/proto/user") {
		t.Fatalf("unexpected proto dirs: %v", dirs)
	}
}
