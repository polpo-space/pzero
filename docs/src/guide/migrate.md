---
title: Database version automatic migration
icon: /icons/carbon-migrate.svg
star: true
order: 5.5
---

* pzero implements database migration capability based on [migrate](https://github.com/golang-migrate/migrate)
* pzero detects files under desc/sql_migration directory by default, executes migration
* Refer to [best practices](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md) on how to write database migration files

## Configuration

```yaml
migrate:
  driver: "mysql"
  datasource-url: "root:123456@tcp(127.0.0.1:3306)/jzero-admin"
```

## Upgrade

```shell
# Upgrade to latest by default
pzero migrate up
# Upgrade n migrations
pzero migrate up 3
```

## Rollback

```shell
# Rollback all by default
pzero migrate down
# Rollback n migrations
pzero migrate down 3
```

## Get version

```shell
pzero migrate version
```

## Force rollback to specific version

```shell
pzero migrate goto <your_version>
```
