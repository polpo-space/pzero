package modelx

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/polpo-space/pzero/core/stores/cache"
)

// MustNewConn opens and checks a PostgreSQL connection without changing global SQL settings.
// An empty driver name and the postgres alias both select pgx. Other databases are unsupported.
// Config files must explicitly set DriverName: pgx: go-zero's config loader defaults it to mysql.
func MustNewConn(c sqlx.SqlConf) sqlx.SqlConn {
	switch c.DriverName {
	case "", "postgres", "pgx":
		c.DriverName = "pgx"
	case "mysql":
		panic("modelx: only PostgreSQL is supported; set DriverName: pgx explicitly (go-zero defaults an omitted DriverName to mysql)")
	default:
		panic(fmt.Errorf("modelx: only PostgreSQL is supported, got driver %q", c.DriverName))
	}
	sqlConn := sqlx.MustNewConn(c)
	db, err := sqlConn.RawDB()
	logx.Must(err)
	err = db.Ping()
	logx.Must(err)

	return sqlConn
}

// NewConnWithCache returns a CachedConn with a custom cache.
func NewConnWithCache(db sqlx.SqlConn, c cache.Cache) sqlc.CachedConn {
	return sqlc.NewConnWithCache(db, c)
}
