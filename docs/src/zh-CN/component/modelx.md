---
title: modelx(数据库连接)
icon: /icons/oui-vis-query-sql.svg
order: 2
---

## 特性

* 仅支持 PostgreSQL，内置 pgx 驱动，无需手动导入。
* `driverName` 可为 `pgx`、`postgres` 或空值，统一使用 pgx；其他值会在启动时失败。
* 不修改进程级 `sqlbuilder.DefaultFlavor`，model 和 condition 固定使用 PostgreSQL。

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
    datasource: "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"
    driverName: "pgx"
```
:::