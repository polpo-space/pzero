package genapi

import (
	"os"
	"path/filepath"
	"testing"
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
