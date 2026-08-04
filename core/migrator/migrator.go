// Package migrator provides PostgreSQL schema migration execution and service
// commands backed by pgx.
package migrator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/eddieowens/opts"
	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	DefaultXMigrationsTable = "schema_migrations"
	DefaultSource           = "file://desc/sql_migration"
)

type (
	Migrator interface {
		// Up looks at the currently active migration version
		// and will migrate all the way up (default applying all up migrations).
		Up(steps ...uint) error

		// Down looks at the currently active migration version
		// and will migrate all the way down (default applying all down migrations).
		Down(steps ...uint) error

		// Goto looks at the currently active migration version,
		// then migrates either up or down to the specified version.
		Goto(version uint) error

		// Force sets a migration version and clears the dirty state without
		// running migrations.
		Force(version int) error

		// Version returns the currently active migration version.
		// If no migration has been applied, yet, it will return ErrNilVersion.
		Version() (version uint, dirty bool, err error)

		// Close source and database, return source error and database error
		Close() (error, error)
	}

	Options struct {
		Source             string
		SourceAppendDriver bool
		XMigrationsTable   string
	}

	defaultMigrator struct {
		migrate *migrate.Migrate
	}
)

func (d *defaultMigrator) Close() (error, error) {
	return d.migrate.Close()
}

func WithSource(source string) opts.Opt[Options] {
	return func(d *Options) {
		d.Source = source
	}
}

func WithSourceAppendDriver(sourceAppendDriver bool) opts.Opt[Options] {
	return func(d *Options) {
		d.SourceAppendDriver = sourceAppendDriver
	}
}

func WithXMigrationsTable(xMigrationsTable string) opts.Opt[Options] {
	return func(u *Options) {
		u.XMigrationsTable = xMigrationsTable
	}
}

func (d Options) DefaultOptions() Options {
	return Options{
		Source:             DefaultSource,
		XMigrationsTable:   DefaultXMigrationsTable,
		SourceAppendDriver: false,
	}
}

func New(sqlConf sqlx.SqlConf, op ...opts.Opt[Options]) (Migrator, error) {
	ops := opts.DefaultApply(op...)

	connConfig, err := parsePgxConfig(sqlConf)
	if err != nil {
		return nil, err
	}

	source := ops.Source
	if ops.SourceAppendDriver {
		source += "/pgx"
	}

	db := stdlib.OpenDB(*connConfig)
	databaseDriver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{
		MigrationsTable: ops.XMigrationsTable,
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(source, "pgx5", databaseDriver)
	if err != nil {
		_ = databaseDriver.Close()
		return nil, err
	}

	return &defaultMigrator{
		migrate: m,
	}, nil
}

func parsePgxConfig(sqlConf sqlx.SqlConf) (*pgx.ConnConfig, error) {
	if sqlConf.DriverName != "pgx" {
		return nil, fmt.Errorf("unsupported migration driver %q: only pgx is supported", sqlConf.DriverName)
	}

	dataSource := sqlConf.DataSource
	for _, scheme := range []string{"pgx5://", "pgx://"} {
		if strings.HasPrefix(dataSource, scheme) {
			dataSource = "postgres://" + strings.TrimPrefix(dataSource, scheme)
			break
		}
	}

	config, err := pgx.ParseConfig(dataSource)
	if err != nil {
		return nil, fmt.Errorf("parse pgx data source: %w", err)
	}
	return config, nil
}

func (d *defaultMigrator) Up(steps ...uint) error {
	if len(steps) > 1 {
		return errors.New("steps number should not be more than 1")
	}

	var err error

	if len(steps) == 0 {
		err = d.migrate.Up()
	} else {
		err = d.migrate.Steps(int(steps[0]))
	}

	return ignoreNoChange(err)
}

func (d *defaultMigrator) Down(steps ...uint) error {
	if len(steps) > 1 {
		return errors.New("steps number should not be more than 1")
	}

	var err error
	if len(steps) == 0 {
		err = d.migrate.Down()
	} else {
		err = d.migrate.Steps(-int(steps[0]))
	}

	return ignoreNoChange(err)
}

func ignoreNoChange(err error) error {
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return err
}

func (d *defaultMigrator) Goto(version uint) error {
	return d.migrate.Migrate(version)
}

func (d *defaultMigrator) Force(version int) error {
	return d.migrate.Force(version)
}

func (d *defaultMigrator) Version() (uint, bool, error) {
	return d.migrate.Version()
}
