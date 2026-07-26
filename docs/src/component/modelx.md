---
title: modelx(Database connection)
icon: /icons/oui-vis-query-sql.svg
order: 2
---

## Features

* Adapts to mysql/postgres/sqlite, no need to import drivers

::: code-tabs#shell

@tab main.go

```go
package main

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/polpo-space/pzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Config struct {
	Sqlx sqlx.SqlConf
}

func main() {
	var c Config
	conf.MustLoad("etc/etc.yaml", &c, conf.UseEnv())

	sqlConn := modelx.MustNewConn(c.Sqlx)

	// connect database
	sqlConn := modelx.MustNewConn(c.Sqlx)

	// execute sql
	result, err := sqlConn.ExecCtx(context.Background(), "select 1")
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

```

@tab etc/etc.yaml
```yaml
sqlx:
    datasource: "jzero-admin.db"
    driverName: "sqlite"
```
:::
