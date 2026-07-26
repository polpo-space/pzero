---
title: modelx(数据库连接)
icon: /icons/oui-vis-query-sql.svg
order: 2
---

## 特性

* 适配 mysql/postgres/sqlite, 无需导入驱动

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

	// 连接数据库
	sqlConn := modelx.MustNewConn(c.Sqlx)

	// 执行 sql
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