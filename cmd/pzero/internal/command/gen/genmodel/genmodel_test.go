package genmodel

import (
	"testing"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
)

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
