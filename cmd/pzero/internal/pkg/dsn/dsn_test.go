package dsn

import (
	"testing"
)

func TestParseDSN(t *testing.T) {
	t.Run("Test ParseDSN with postgres", func(t *testing.T) {
		t.Run("Test ParseDSN with postgres and valid dsn", func(t *testing.T) {
			meta, err := ParseDSN("postgres", "postgres://user:password@localhost:5432/dbname")
			if err != nil {
				t.Errorf("ParseDSN() error = %v", err)
			}
			if meta[User] != "user" {
				t.Errorf("ParseDSN() user = %v, want %v", meta[User], "user")
			}
			if meta[Host] != "localhost" {
				t.Errorf("ParseDSN() host = %v, want %v", meta[Host], "localhost")
			}
			if meta[Port] != "5432" {
				t.Errorf("ParseDSN() port = %v, want %v", meta[Port], "5432")
			}
			if meta[Database] != "dbname" {
				t.Errorf("ParseDSN() dbname = %v, want %v", meta[Database], "dbname")
			}
		})
	})

	t.Run("Test ParseDSN with unsupported driver", func(t *testing.T) {
		_, err := ParseDSN("mysql", "user:password@tcp(localhost:3306)/dbname")
		if err == nil {
			t.Fatal("ParseDSN() expected error for unsupported driver")
		}
	})
}
