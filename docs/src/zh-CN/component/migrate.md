---
title: migrate(管理数据库迁移脚本)
icon: /icons/streamline-plump-color-database.svg
order: 3
---

migrate 组件从 `desc/sql_migration` 读取并管理 PostgreSQL migration，
只支持 go-zero 的 `pgx` driver。

* Up：应用全部待执行 migration，也可以限制执行步数
* Down：回滚 migration，也可以限制回滚步数
* Goto：迁移到指定版本
* Force：不执行 migration，直接设置版本并清除 dirty 状态
* Version：获取当前版本和 dirty 状态

::: code-tabs#shell

@tab main.go

```go
package main

import (
	"github.com/polpo-space/pzero/runtime/migrator"
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

默认直接读取 `desc/sql_migration`。如果现有项目仍按 driver 保存 migration，
可以传入 `migrate.WithSourceAppendDriver(true)`，读取 `desc/sql_migration/pgx`。
