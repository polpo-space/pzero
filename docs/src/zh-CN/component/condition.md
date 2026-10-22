---
title: condition(数据库条件查询)
icon: /icons/material-symbols-conditions.svg
order: 4
---

condition 的核心在于构造出 statement 和 args 参数, 然后搭配 go-zero 的 sqlx 执行器实际执行.

## 特性

* 基于 go-sqlbuilder 构造 PostgreSQL SQL，统一使用双引号和 `$1`、`$2` 占位符，无需设置全局方言
* 支持链式调用方便使用

:::tip 搭配 pzero 的数据库代码自动生成的功能, 仅需构造出 conditions 即可
:::

## 查询场景

::: code-tabs#shell

@tab condition

```go
package main

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/polpo-space/pzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/huandu/go-sqlbuilder"
	"github.com/polpo-space/pzero/core/stores/condition"
)

type Config struct {
	Sqlx sqlx.SqlConf
}

func main() {
	// 加载配置
	var c Config
	conf.MustLoad("etc/etc.yaml", &c, conf.UseEnv())

	// 连接 PostgreSQL
	sqlConn := modelx.MustNewConn(c.Sqlx)

	conditions := condition.New(condition.Condition{
		// 操作的字段
		Field: "name",
		// 操作
		Operator: condition.Equal,
		// 字段的值
		Value: "pzero",
		// ValueFunc 优先级比 Skip 高
		ValueFunc: func() any {
			return "pzero"
		},
		// 是否跳过该条件
		Skip: false,
		// SkipFunc 优先级比 Skip 高
		SkipFunc: func() bool {
			return false
		},
	})

	statement, args := condition.BuildSelect(sqlbuilder.Select("*").From("user"), conditions...)
	fmt.Println(statement, args)

	type User struct {
		Id   int64  `db:"id"`
		Name string `db:"name"`
	}

	var users []User

	err := sqlConn.QueryRowsCtx(context.Background(), &users, statement, args...)
	if err != nil {
		panic(err)
	}

	fmt.Println(users)
}

```

@tab condition chain

```go
package main

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/polpo-space/pzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/huandu/go-sqlbuilder"
	"github.com/polpo-space/pzero/core/stores/condition"
)

type Config struct {
	Sqlx sqlx.SqlConf
}

func main() {
	// 加载配置
	var c Config
	conf.MustLoad("etc/etc.yaml", &c, conf.UseEnv())

	// 连接 PostgreSQL
	sqlConn := modelx.MustNewConn(c.Sqlx)

	chain := condition.NewChain().Equal("name", "pzero",
		// WithValueFunc 比 value 优先级高
		condition.WithValueFunc(func() any {
			return "pzero"
		}),
		// 是否跳过该条件
		condition.WithSkip(false),
		// WithSkipFunc 优先级比 WithSkip 高
		condition.WithSkipFunc(
			func() bool {
				return false
			}),
	)

	statement, args := condition.BuildSelect(sqlbuilder.Select("*").From("user"), chain.Build()...)
	fmt.Println(statement, args)

	type User struct {
		Id   int64  `db:"id"`
		Name string `db:"name"`
	}

	var users []User

	err := sqlConn.QueryRowsCtx(context.Background(), &users, statement, args...)
	if err != nil {
		panic(err)
	}

	fmt.Println(users)
}
```

@tab etc/etc.yaml

```yaml
sqlx:
    datasource: "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"
    driverName: "pgx"
```

:::

## 更新场景

::: code-tabs#shell

@tab condition

```go
package main

import (
	"context"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/polpo-space/pzero/core/stores/condition"
	"github.com/polpo-space/pzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Config struct {
	Sqlx sqlx.SqlConf
}

func main() {
	// 加载配置
	var c Config
	conf.MustLoad("etc/etc.yaml", &c, conf.UseEnv())

	// 连接 PostgreSQL
	sqlConn := modelx.MustNewConn(c.Sqlx)

	conditions := condition.New(condition.Condition{
		// 操作的字段
		Field: "name",
		// 操作
		Operator: condition.Equal,
		// 字段的值
		Value: "pzero",
		// ValueFunc 优先级比 Skip 高
		ValueFunc: func() any {
			return "pzero"
		},
		// 是否跳过该条件
		Skip: false,
		// SkipFunc 优先级比 Skip 高
		SkipFunc: func() bool {
			return false
		},
	})

	statement, args := condition.BuildUpdate(
		sqlbuilder.Update("user"),
		// 设置更新字段, 直接使用 map
		map[string]any{
			"name": "pzero",
			"version": condition.UpdateField{
				Operator: condition.Incr,
			},
		},
		conditions...)

	result, err := sqlConn.ExecCtx(context.Background(), statement, args...)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
```

@tab condition chain

```go
package main

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/polpo-space/pzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/huandu/go-sqlbuilder"
	"github.com/polpo-space/pzero/core/stores/condition"
)

type Config struct {
	Sqlx sqlx.SqlConf
}

func main() {
	// 加载配置
	var c Config
	conf.MustLoad("etc/etc.yaml", &c, conf.UseEnv())

	// 连接 PostgreSQL
	sqlConn := modelx.MustNewConn(c.Sqlx)

	chain := condition.NewChain().Equal("name", "pzero",
		// WithValueFunc 比 value 优先级高
		condition.WithValueFunc(func() any {
			return "pzero"
		}),
		// 是否跳过该条件
		condition.WithSkip(false),
		// WithSkipFunc 优先级比 WithSkip 高
		condition.WithSkipFunc(
			func() bool {
				return false
			}),
	)

	statement, args := condition.BuildUpdate(
		sqlbuilder.Update("user"),
		// 设置更新字段, 构造出 map
		condition.NewUpdateFieldChain().
			Assign("name", "pzero").
			Incr("version").
			Build(),
		chain.Build()...)
	fmt.Println(statement, args)

	result, err := sqlConn.ExecCtx(context.Background(), statement, args...)
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

## 删除场景

::: code-tabs#shell

@tab condition

```go
package main

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/polpo-space/pzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/huandu/go-sqlbuilder"
	"github.com/polpo-space/pzero/core/stores/condition"
)

type Config struct {
	Sqlx sqlx.SqlConf
}

func main() {
	// 加载配置
	var c Config
	conf.MustLoad("etc/etc.yaml", &c, conf.UseEnv())

	// 连接 PostgreSQL
	sqlConn := modelx.MustNewConn(c.Sqlx)

	conditions := condition.New(condition.Condition{
		// 操作的字段
		Field: "name",
		// 操作
		Operator: condition.Equal,
		// 字段的值
		Value: "pzero",
		// ValueFunc 优先级比 Skip 高
		ValueFunc: func() any {
			return "pzero"
		},
		// 是否跳过该条件
		Skip: false,
		// SkipFunc 优先级比 Skip 高
		SkipFunc: func() bool {
			return false
		},
	})

	statement, args := condition.BuildDelete(sqlbuilder.DeleteFrom("user"), conditions...)
	fmt.Println(statement, args)

	result, err := sqlConn.ExecCtx(context.Background(), statement, args...)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

```

@tab condition chain

```go
package main

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/polpo-space/pzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/huandu/go-sqlbuilder"
	"github.com/polpo-space/pzero/core/stores/condition"
)

type Config struct {
	Sqlx sqlx.SqlConf
}

func main() {
	// 加载配置
	var c Config
	conf.MustLoad("etc/etc.yaml", &c, conf.UseEnv())

	// 连接 PostgreSQL
	sqlConn := modelx.MustNewConn(c.Sqlx)

	chain := condition.NewChain().Equal("name", "pzero",
		// WithValueFunc 比 value 优先级高
		condition.WithValueFunc(func() any {
			return "pzero"
		}),
		// 是否跳过该条件
		condition.WithSkip(false),
		// WithSkipFunc 优先级比 WithSkip 高
		condition.WithSkipFunc(
			func() bool {
				return false
			}),
	)

	statement, args := condition.BuildDelete(sqlbuilder.DeleteFrom("user"), chain.Build()...)
	fmt.Println(statement, args)

	result, err := sqlConn.ExecCtx(context.Background(), statement, args...)
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