package modelx

import (
	"fmt"
	"strings"
	"testing"

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
