package modelx

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestMustNewConnRejectsUnsupportedDrivers(t *testing.T) {
	for _, driver := range []string{"mysql", "sqlite", "unknown"} {
		t.Run(driver, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered == nil || !strings.Contains(fmt.Sprint(recovered), "only PostgreSQL is supported") {
					t.Fatalf("expected unsupported driver failure, got %v", recovered)
				}
			}()
			MustNewConn(sqlx.SqlConf{DriverName: driver, DataSource: "unused"})
		})
	}
}

func TestLoadedConfigRequiresExplicitPostgresDriver(t *testing.T) {
	var config struct{ Sqlx sqlx.SqlConf }
	if err := conf.LoadFromYamlBytes([]byte("Sqlx:\n  DataSource: postgres://localhost/test\n"), &config); err != nil {
		t.Fatal(err)
	}
	if config.Sqlx.DriverName != "mysql" {
		t.Fatalf("unexpected upstream default: %q", config.Sqlx.DriverName)
	}
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(fmt.Sprint(r), "DriverName: pgx") {
			t.Fatalf("missing actionable driver error: %v", r)
		}
	}()
	MustNewConn(config.Sqlx)
}
