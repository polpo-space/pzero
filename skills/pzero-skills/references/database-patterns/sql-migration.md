# SQL Migration Guide

## Runtime contract

- API and RPC projects generated with the `model` feature include a `migrate` subcommand.
- Migration execution lives in `github.com/polpo-space/pzero/runtime/migrator`.
- The generated service `migrate` cobra command lives in `github.com/polpo-space/pzero/runtime/migrator/cmd`.
- Migration execution supports PostgreSQL through `pgx` only.
- Commands load `sqlx.SqlConf` from the service root's inherited `--config` flag.
- Starting the API or RPC server never applies migrations automatically.

Run migrations through the generated service binary:

```bash
go run . migrate create add_users
go run . migrate up --config etc/etc.yaml
go run . migrate down --steps 1 --config etc/etc.yaml
go run . migrate goto 3 --config etc/etc.yaml
go run . migrate version --config etc/etc.yaml
```

Use `force` only after manually verifying the database state:

```bash
go run . migrate force 3 --config etc/etc.yaml
```

`create` writes a Unix-timestamped `.up.sql` and `.down.sql` pair without loading database configuration.

## Configuration

Use a pgx SQL configuration:

```yaml
sqlx:
  driverName: pgx
  dataSource: postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable
```

MySQL, SQLite, and empty driver names are rejected before a migration connection is opened.

## Rules

1. Every schema change needs migrations
2. Always provide both `up.sql` and `down.sql`
3. Apply migrations as an explicit development, release, or operator action
4. Never couple migrations to API or RPC server startup

## Example

`migrate create` names files with a Unix timestamp version prefix:

```text
desc/sql_migration/
├── 1748252730_create_users_table.up.sql
├── 1748252730_create_users_table.down.sql
├── 1748256330_add_email_index.up.sql
└── 1748256330_add_email_index.down.sql
```

## Guidance

- Never rewrite an old shipped migration
- Keep one logical change per migration
- Test rollbacks
- Keep migration versions unique; generated files use Unix timestamps
- Commit the migration pair before regenerating models
