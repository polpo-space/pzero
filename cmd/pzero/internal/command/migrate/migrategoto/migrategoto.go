package migrategoto

import (
	"errors"

	"github.com/polpo-space/pzero/core/migrator"
	"github.com/spf13/cast"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
)

func Run(args []string) error {
	m, err := migrator.New(sqlx.SqlConf{
		DataSource: config.C.Migrate.DataSourceUrl,
		DriverName: config.C.Migrate.Driver,
	},
		migrator.WithXMigrationsTable(config.C.Migrate.XMigrationsTable),
		migrator.WithSource(config.C.Migrate.Source),
		migrator.WithSourceAppendDriver(config.C.Migrate.SourceAppendDriver))
	if err != nil {
		return err
	}

	if len(args) <= 0 {
		return errors.New("step must be greater than 0")
	}
	return m.Goto(cast.ToUint(args[0]))
}
