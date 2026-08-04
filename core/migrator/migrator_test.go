package migrator

import (
	"errors"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestNewRejectsNonPgxDrivers(t *testing.T) {
	t.Parallel()

	for _, driver := range []string{"mysql", "sqlite", ""} {
		t.Run(driver, func(t *testing.T) {
			_, err := New(sqlx.SqlConf{
				DriverName: driver,
				DataSource: "unused",
			}, WithSource("file://"+t.TempDir()))
			if err == nil {
				t.Fatalf("expected driver %q to be rejected", driver)
			}
			if !strings.Contains(err.Error(), "only pgx is supported") {
				t.Fatalf("unexpected error for driver %q: %v", driver, err)
			}
		})
	}
}

func TestParsePgxConfig(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		dataSource string
	}{
		{
			name:       "postgres scheme",
			dataSource: "postgres://postgres:secret@localhost:5432/app?sslmode=disable",
		},
		{
			name:       "postgresql scheme",
			dataSource: "postgresql://postgres:secret@localhost:5432/app?sslmode=disable",
		},
		{
			name:       "libpq key value",
			dataSource: "host=localhost port=5432 user=postgres password=secret dbname=app sslmode=disable",
		},
		{
			name:       "legacy pgx5 scheme",
			dataSource: "pgx5://postgres:secret@localhost:5432/app?sslmode=disable",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			config, err := parsePgxConfig(sqlx.SqlConf{
				DriverName: "pgx",
				DataSource: test.dataSource,
			})
			if err != nil {
				t.Fatalf("parsePgxConfig() error: %v", err)
			}
			if config.Database != "app" {
				t.Fatalf("parsePgxConfig() database = %q, want app", config.Database)
			}
		})
	}
}

func TestParsePgxConfigRejectsInvalidDataSource(t *testing.T) {
	t.Parallel()

	_, err := parsePgxConfig(sqlx.SqlConf{DriverName: "pgx", DataSource: "://invalid"})
	if err == nil || !strings.Contains(err.Error(), "parse pgx data source") {
		t.Fatalf("expected pgx parse error, got %v", err)
	}
}

func TestIgnoreNoChange(t *testing.T) {
	t.Parallel()

	if err := ignoreNoChange(migrate.ErrNoChange); err != nil {
		t.Fatalf("expected ErrNoChange to be ignored, got %v", err)
	}
	expected := errors.New("migration failed")
	if err := ignoreNoChange(expected); !errors.Is(err, expected) {
		t.Fatalf("expected original error, got %v", err)
	}
}
