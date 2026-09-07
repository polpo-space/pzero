---
title: Database migration
icon: /icons/carbon-migrate.svg
star: true
order: 5.5
---

* API and RPC projects generated with the `model` feature include a `migrate` subcommand
* Migrations are loaded from `desc/sql_migration` and executed explicitly through the generated service binary
* Migration execution supports PostgreSQL through the `pgx` driver only
* Starting the service never applies migrations automatically
* Refer to [best practices](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md) for writing migration files

## Configuration

```yaml
sqlx:
  driverName: "pgx"
  dataSource: "postgres://postgres:postgres@127.0.0.1:5432/jzero-admin?sslmode=disable"
```

## Create

```shell
go run . migrate create add_users
```

## Upgrade and rollback

```shell
# Apply all pending migrations
go run . migrate up --config etc/etc.yaml
# Apply at most three migrations
go run . migrate up --steps 3 --config etc/etc.yaml
# Roll back one migration
go run . migrate down --steps 1 --config etc/etc.yaml
```

## Version control

```shell
go run . migrate version --config etc/etc.yaml
go run . migrate goto <version> --config etc/etc.yaml
# Clear dirty state only after manually verifying the database
go run . migrate force <version> --config etc/etc.yaml
```
