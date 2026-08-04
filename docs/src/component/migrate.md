---
title: migrate(Database migration management)
icon: /icons/streamline-plump-color-database.svg
order: 3
---

The migrate component manages PostgreSQL migrations from `desc/sql_migration`.
Only the go-zero `pgx` driver is supported.

* Up: apply all pending migrations, or a limited number of steps
* Down: roll back all migrations, or a limited number of steps
* Goto: migrate to a specific version
* Force: set a version and clear dirty state without running migrations
* Version: get the current version and dirty state

::: code-tabs#shell

@tab main.go

```go
package main

import (
	"github.com/polpo-space/pzero/core/migrator"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Config struct {
	Sqlx sqlx.SqlConf
}

func main() {
	var c Config
	conf.MustLoad("etc/etc.yaml", &c, conf.UseEnv())

	m, err := migrator.New(c.Sqlx)
	if err != nil {
		panic(err)
	}
	defer m.Close()

	if err = m.Up(); err != nil {
		panic(err)
	}
}
```

@tab etc/etc.yaml

```yaml
sqlx:
  driverName: "pgx"
  dataSource: "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"
```

@tab desc/sql_migration/1_init.up.sql

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

@tab desc/sql_migration/1_init.down.sql

```sql
DROP TABLE IF EXISTS users;
```

:::

By default, migrations are read directly from `desc/sql_migration`.
For an existing layout that keeps PostgreSQL migrations in a driver subdirectory,
use `migrate.WithSourceAppendDriver(true)` to read `desc/sql_migration/pgx`.
