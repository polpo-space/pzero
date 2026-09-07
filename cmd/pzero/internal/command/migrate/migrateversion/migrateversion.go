package migrateversion

import (
	"fmt"

	"github.com/polpo-space/pzero/runtime/migrator"
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

	version, dirty, err := m.Version()
	if err != nil {
		return err
	}

	if dirty {
		fmt.Printf("%v (dirty)\n", version)
	} else {
		fmt.Printf("%v\n", version)
	}
	return nil
}
