package migrator

import "github.com/eddieowens/opts"

var (
	DefaultXMigrationsTable = "schema_migrations"
	DefaultSource           = "file://desc/sql_migration"
)

type Options struct {
	Source             string
	SourceAppendDriver bool
	XMigrationsTable   string
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
