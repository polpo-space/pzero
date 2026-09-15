package genmodel

import (
	"os"
	"strings"
	"testing"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
)

func TestRejectMultiDatasourceBeforeWriting(t *testing.T) {
	orig := config.C
	t.Cleanup(func() { config.C = orig })
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("internal/model", 0o755); err != nil {
		t.Fatal(err)
	}
	const existing = "existing model registration\n"
	if err := os.WriteFile("internal/model/model.go", []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, want   string
		urls, tables []string
	}{
		{"multiple URLs", "exactly one", []string{"postgres://localhost/a", "postgres://localhost/b"}, []string{"*"}},
		{"blank URL", "must not be empty", []string{" "}, []string{"users"}},
		{"qualified table", "qualified table", []string{"postgres://localhost/a"}, []string{"a.users"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config.C = config.Config{Gen: config.GenConfig{ModelDatasource: true, ModelDatasourceUrl: tc.urls, ModelDatasourceTable: tc.tables}}
			_, err := (&PzeroModel{Module: "example.com/app"}).Gen(nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
			data, err := os.ReadFile("internal/model/model.go")
			if err != nil || string(data) != existing {
				t.Fatalf("rejected input changed registration: %q, %v", data, err)
			}
		})
	}
}

func TestHasExplicitSQLDesc(t *testing.T) {
	orig := config.C
	t.Cleanup(func() { config.C = orig })

	config.C = config.Config{}

	t.Run("empty desc", func(t *testing.T) {
		config.C.Gen.Desc = nil
		if hasExplicitSQLDesc() {
			t.Fatal("expected false when desc is empty")
		}
	})

	t.Run("api desc only", func(t *testing.T) {
		config.C.Gen.Desc = []string{"desc/api/render.api"}
		if hasExplicitSQLDesc() {
			t.Fatal("api desc must not be treated as sql input")
		}
	})

	t.Run("sql file desc", func(t *testing.T) {
		config.C.Gen.Desc = []string{"desc/sql/users.sql"}
		if !hasExplicitSQLDesc() {
			t.Fatal("sql file via --desc must be detected")
		}
	})

	t.Run("sql dir desc", func(t *testing.T) {
		config.C.Gen.Desc = []string{"desc/sql"}
		if !hasExplicitSQLDesc() {
			t.Fatal("desc/sql directory via --desc must be detected")
		}
	})

	t.Run("sql dir with dot slash", func(t *testing.T) {
		config.C.Gen.Desc = []string{"./desc/sql"}
		if !hasExplicitSQLDesc() {
			t.Fatal("./desc/sql must match desc/sql after Abs")
		}
	})

	t.Run("sql backup prefix sibling", func(t *testing.T) {
		config.C.Gen.Desc = []string{"desc/sql_backup"}
		if hasExplicitSQLDesc() {
			t.Fatal("desc/sql_backup must not be treated as desc/sql")
		}
	})
}

func TestGenRequiresURLEvenWhenAPIDescSet(t *testing.T) {
	orig := config.C
	t.Cleanup(func() { config.C = orig })

	config.C = config.Config{}
	config.C.Gen.ModelDatasource = true
	config.C.Gen.Desc = []string{"desc/api/user.api"}

	_, err := (&PzeroModel{Module: "example.com/app"}).Gen(nil)
	if err == nil {
		t.Fatal("expected url required error when desc is api-only")
	}
	if err.Error() != "model-datasource-url is required when model-datasource is enabled" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenNoopWhenDatasourceDisabled(t *testing.T) {
	orig := config.C
	t.Cleanup(func() { config.C = orig })

	config.C = config.Config{}
	config.C.Gen.Desc = []string{"desc/api/user.api"}

	files, err := (&PzeroModel{Module: "example.com/app"}).Gen(nil)
	if err != nil || files != nil {
		t.Fatalf("disabled datasource must stay noop, files=%v err=%v", files, err)
	}
}

func TestNormalizeModelDriver(t *testing.T) {
	got, err := normalizeModelDriver("postgres")
	if err != nil || got != "pgx" {
		t.Fatalf("postgres => pgx, got %q err=%v", got, err)
	}
	got, err = normalizeModelDriver("")
	if err != nil || got != "pgx" {
		t.Fatalf("empty => pgx, got %q err=%v", got, err)
	}
	if _, err := normalizeModelDriver("mysql"); err == nil {
		t.Fatal("mysql must be rejected")
	}
}
