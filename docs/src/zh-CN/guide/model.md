---
title: model 指南
icon: /icons/uil-database.svg
star: true
order: 4
---

## 前言

pzero 通过 PostgreSQL 数据源生成数据库代码到 `internal/model` 下。

为了在使用上更加方便，pzero 自动生成 `internal/model/model.go` 文件，用于注册所有生成的模型代码。

当前 model 生成器已收缩为 PostgreSQL-only：
* `gen.model-driver` 支持 `postgres`（推荐）或 `pgx`
* model 生成只支持远程数据源模式（`model-datasource: true`）
* `desc/sql` 是 **schema snapshot**（当前结构镜像），可与 datasource 共存；**不会**、也**不能**作为 `pzero gen` 的 model 输入
* 通过 `--desc` / `gen.desc` 指定 api/proto 时，全量 `pzero gen` 会跳过 model 生成（限定生成范围）；不要把 `.sql` 传给 `--desc`
* 只要刷表 model、不重写 api/rpc 时，使用 `pzero gen model`

## Schema 三分法

| 路径 | 职责 |
| --- | --- |
| `desc/sql_migration/` | 正式 schema 变更（唯一演进入口） |
| `desc/sql/` | 当前结构 snapshot，供评审/对齐，不喂 gen |
| datasource + `model-datasource-table` | model 代码生成的唯一输入 |

`internal/model/model.go` 由 `model-datasource-table`（或 `*`）全量重写。表列表必须覆盖所有需要注册的表；手维额外注册会在下次 gen 时被覆盖。

## 特性

* 支持多数据源生成代码
* 支持 redis/自定义缓存
* 单数据源和多 schema 项目生成出的模型注册方式保持一致

## 缓存配置

pzero 支持灵活的模型缓存配置:

* **model-cache**: 启用或禁用模型缓存(默认: false)
* **model-cache-table**: 指定哪些表需要缓存(默认: * 表示所有表)
* **model-cache-expiry-table**: 为特定表配置自定义缓存过期时间

### 缓存过期时间配置

`model-cache-expiry-table` 选项允许为不同的表设置不同的缓存过期时间:

```yaml
gen:
  model-cache-expiry-table:
    - table: manage_user
      expiry: 3600              # 查询到的数据缓存时间(秒)
      not-found-expiry: 60      # 未查询到数据的缓存时间(秒)
    - table: manage_role
      expiry: 7200
      not-found-expiry: 120
```

**字段说明:**
* **table**: 要应用缓存过期设置的表名
* **expiry**: 成功查询到的数据缓存时长,单位秒(默认: 系统默认值)
* **not-found-expiry**: 未查询到数据时的缓存时长,单位秒(默认: 系统默认值)

该配置特别适用于:
* 数据更新频率不同的表
* 减少频繁访问的参考数据的缓存未命中
* 针对特定业务场景优化缓存命中率

## 基于远程数据源地址生成代码

```yaml
gen:
  model-driver: postgres
  # 是否生成带缓存的数据库代码
  model-cache: true
  # 缓存表, 默认为 *(所有)
  model-cache-table:
    - manage_user
  # 为指定表设置缓存过期时间(单位: 秒)
  model-cache-expiry-table:
    - table: manage_user
      expiry: 3600
      not-found-expiry: 60
  # PostgreSQL model 生成必须开启 datasource 模式
  model-datasource: true
  # PostgreSQL 数据源配置
  model-datasource-url: "postgres://postgres:postgres@127.0.0.1:5432/jzero-admin?sslmode=disable"
  # schema, 默认为 public
  model-schema: public
  # Ignore columns while creating or updating rows, 默认为 create_at,created_at,create_time,update_at,updated_at,update_time
  model-ignore-columns: ["create_time", "update_time"]
  # 使用哪些 table, 默认为 *(所有)
  model-datasource-table:
    - manage_email
    - manage_menu
    - manage_role
    - manage_role_menu
    - manage_user
    - manage_user_role
```


```shell
pzero gen
pzero gen model
```

`pzero gen` 仍会跑 model → api → rpc。`pzero gen model` 只写所选表的 `internal/model`（含 `model.go` 注册）。

pzero 支持多数据源。通过在 `model-datasource-table` 中使用 `database.table` 形式，可以将表映射到指定的数据源。

例如增加 `jzero-admin_log.operate_log`：

```yaml
gen:
  model-driver: postgres
  model-datasource: true
  model-datasource-url:
    - "postgres://postgres:postgres@127.0.0.1:5432/jzero-admin?sslmode=disable"
    - "postgres://postgres:postgres@127.0.0.1:5432/jzero-admin_log?sslmode=disable"
  model-schema: public
  model-ignore-columns: ["create_time", "update_time"]
  model-datasource-table:
    - manage_email
    - manage_menu
    - manage_role
    - manage_role_menu
    - manage_user
    - manage_user_role
    - jzero-admin_log.operate_log
```

```shell
pzero gen
pzero gen model
```

生成的 `internal/model/model.go` 会与所选 PostgreSQL 数据源集合保持一致。

## 默认生成如下方法

* Insert: 插入单条数据（PostgreSQL 通过 `RETURNING` 回填自增主键）
* BulkInsert: 批量插入数据
* Update: 根据主键更新单条数据
* Delete: 根据主键删除单条数据
* FindOne: 根据主键查询单条数据
* FindOneForUpdate: 在调用方事务内按主键查询并加行锁，跳过缓存
* FindByCondition: 条件查询
* FindFieldsByCondition: 条件查询(指定查询字段)
* FindOneByCondition: 条件查询单条数据（按唯一索引查询请用此方法，不再生成 `FindOneBy*` 系列）
* FindOneFieldsByCondition: 条件查询单条数据(指定查询字段)
* CountByCondition: 条件查询总数
* PageByCondition: 条件分页查询
* UpdateFieldsByCondition: 条件更新指定字段
* UpdateFieldsByConditionAffected: 条件更新指定字段并返回受影响行数
* DeleteByCondition: 条件删除

具体使用请参阅: [condition组件](../component/condition.md)

## PostgreSQL 数组与类型覆盖

数组字段可以直接参与生成，不需要为 `TEXT[]` 跳过整张表。生成的 model 使用 `github.com/lib/pq` 的数组类型，支持 `database/sql.Scanner` 与 `driver.Valuer`；连接驱动仍可使用 `pgx`。

| PostgreSQL 类型 | 默认 Go 类型 |
| --- | --- |
| `TEXT[]`、`VARCHAR[]` | `pq.StringArray` |
| `BOOLEAN[]` | `pq.BoolArray` |
| `SMALLINT[]`、`INTEGER[]`、`BIGINT[]` | `pq.Int64Array` |
| `REAL[]`、`DOUBLE PRECISION[]` | `pq.Float64Array` |

例如 `documents.image_urls TEXT[]` 会生成 `ImageUrls pq.StringArray`。`Insert`、`BulkInsert`、`Update` 和查询方法直接使用该字段：

```go
document := &documents.Documents{
    Name: "example",
    ImageUrls: []string{"https://example.com/photo.png"},
}
err := documentModel.Insert(ctx, nil, document)
```

`nil` 数组写入 SQL `NULL`，非 nil 的空数组写入 `{}`，读取后仍保留区别。默认类型面向一维、元素非 NULL 的数组；多维数组、数组中的 NULL 元素等需要自定义类型。通过 `map[string]any` 更新时使用 `pq.StringArray` 等实现 `driver.Valuer` 的类型，不要把普通 `[]string` 当作通用数据库参数。

类型覆盖复用固定版本 goctl 的 `goctl.yaml`。在目标项目的 Go module 根目录配置数据库类型映射：

```yaml
model:
  types_map:
    _text:
      type: types.StringList
      null_type: types.StringList
      pkg: example.com/myapp/internal/types
```

其中 `_text` 是 PostgreSQL catalog 的数组类型名，其他常见键为 `_bool`、`_int8`、`_float8`。`type` 和 `null_type` 分别控制非空、可空列，`pkg` 指定 Go import。映射作用于该数据库类型的所有列，不是逐字段配置。自定义类型需要支持扫描和写入，例如：

```go
package types

import "github.com/lib/pq"

type StringList = pq.StringArray
```

```shell
GOCTL_EXPERIMENTAL=on pzero gen model
go mod tidy
```

`GOCTL_EXPERIMENTAL=on` 启用 goctl 的配置读取；`go mod tidy` 为生成项目补齐 `lib/pq` 或自定义类型的依赖。无需在 `.pzero.yaml` 中再配置一套类型映射。

## 条件更新与事务行锁

`UpdateFieldsByConditionAffected` 返回 `(int64, error)`，可用于比较并更新状态。原 `UpdateFieldsByCondition` 继续只返回 `error`。nil 或空更新 map 明确视为 no-op，返回 `0, nil`；这不是“状态条件未匹配”的证明。要求严格 CAS 的应用应先校验更新字段非空。SQL 执行或读取 `RowsAffected` 失败会返回错误。

```go
affected, err := documentModel.UpdateFieldsByConditionAffected(ctx, nil,
    map[string]any{"status": nextStatus},
    condition.NewChain().
        Equal(documents.Id, documentID).
        Equal(documents.Status, expectedStatus).
        Build()...,
)
if err != nil {
    return err
}
if affected != 1 {
    return errStateConflict // 应用定义的状态冲突
}
```

数据库更新成功但缓存删除失败时，返回真实受影响行数和缓存错误，例如 `(1, err)`。错误不等于“未更新”，不要据此盲目重试；有事务时，最终提交/回滚结果仍由调用方决定。`RowsAffected` 自身失败时行数不可用，应按错误处理。

必须在条件中包含主键和预期状态，受影响行数才表达该条记录的状态竞争结果。方法本身仍允许批量更新，不自动补充业务条件。

需要先读取再修改、并将多个操作放进同一事务时，应用组织事务并传入回调中的 session：

```go
err := conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
    document, err := documentModel.FindOneForUpdate(ctx, session, documentID)
    if err != nil {
        return err
    }
    if document.Status != expectedStatus {
        return errStateConflict
    }
    document.Status = nextStatus
    return documentModel.Update(ctx, session, document)
})
```

此示例适用于禁用 model 缓存的事务写路径。有缓存时，现有写方法会在事务提交前删除缓存；并发普通查询可能在提交前读到旧数据并重新填入缓存。本接口只保证锁定查询绕过缓存，不提供提交后的自动缓存失效。需要强一致性的事务路径应禁用该 model 的缓存，或由应用落实提交后失效和相应的一致性策略。

`FindOneForUpdate` 使用 `SELECT ... FOR UPDATE`，始终读取数据库，不走缓存；传入 nil 或普通 `sqlx.SqlConn` 返回 `modelx.ErrTransactionRequired`，查无记录返回 model 的 `ErrNotFound`。Go 的 `sqlx.Session` 接口不能证明 session 来自真实事务，即使拒绝了明显的普通连接，框架也无法验证任意包装后的 Session。因此只能传入 `TransactCtx` 回调中的原始 session，不能传普通连接或用包装器冒充事务。锁直到调用方提交或回滚才释放；框架不创建隐式事务。

## 生成文件、手写扩展与升级

* `*_model_gen.go` 包含数据类型和通用实现，重新生成时覆盖。
* `*_model.go` 保留业务扩展，已存在时不会被覆盖。生成方法放在嵌入的 `default...Model` 上；已有 `custom...Model` 同名方法可以继续覆盖默认实现，方法签名需与接口一致。
* `internal/model/model.go` 仍按完整表列表重新注册。恢复曾因特殊字段而跳过的表时，应把这些表加入列表，并移除仅为跳过表而维护的注册模板；保留真正的业务扩展。

ORM 仅支持 PostgreSQL。`modelx.MustNewConn` 使用 pgx 驱动，不修改 `sqlbuilder.DefaultFlavor`；生成 model 和 condition 直接构造 PostgreSQL SQL，实例只保存连接、缓存和字段元数据，不再保存可切换的方言。自增主键统一通过 `INSERT ... RETURNING` 回填。

升级时应同步更新 pzero CLI、应用的 pzero runtime 依赖，并重新生成全部 model。项目 `.template/go-zero/model` 下的自定义 model 模板必须整套同步，不能只升级 `types.tpl` 或 `find-one.tpl` 等单个文件；接口、实现、导入和构造模板相互依赖。本次收敛移除了 `ModelOpts.Flavor`、`modelx.WithFlavor`、`MustNewConnAndSqlbuilderFlavor` 和 condition 的 `*WithFlavor` 接口。手写代码与自定义模板按下表迁移：

| 原用法 | PostgreSQL 用法 |
| --- | --- |
| `NewXModel(conn, modelx.WithFlavor(...))` | `NewXModel(conn)`，其余缓存选项保持不变 |
| `MustNewConnAndSqlbuilderFlavor(conf)` | `MustNewConn(conf)` |
| `BuildSelect/Update/DeleteWithFlavor(flavor, ...)` | `BuildSelect/Update/Delete(...)` |
| `QuoteWithFlavor(flavor, name)` | `Quote(name)` |
| `RawFieldNamesWithFlavor`、`RemoveIgnoreColumnsWithFlavor`、`SelectByWhereRawSqlWithFlavor` | 去掉 `WithFlavor` 后缀与首个方言参数 |
| 旧 `Select/Update/DeleteWithFlavor` | 使用 `BuildSelect/Update/Delete` |

应用配置必须显式写 `DriverName: pgx`（或 `postgres`）。go-zero 的 `sqlx.SqlConf` 在 `conf.MustLoad` 等加载流程中会把缺省驱动填为 `mysql`；这与直接构造 Go 结构体的空字符串不同。漏配会在启动时失败并提示补充 `DriverName: pgx`。

```yaml
Sqlx:
  DriverName: pgx
  DataSource: postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable
```

旧生成代码仍可能访问被移除的方言选项或全局状态，不能仅升级 runtime 而不重新生成。

为兼容 PostgreSQL 手写扩展，原包级 `xxxRows` / `xxxFieldNames` 等元数据保留为固定的 PostgreSQL 值，构造 model 时不再改写。手写扩展优先使用实例的 `m.rows`、`m.fieldNames`；不要修改这些元数据。

## 验证生成结果

CI 用独立 PostgreSQL 验证代表性业务表、常用数组、无缓存/有缓存/类型覆盖、条件更新、回滚、行锁竞争和重复生成保留扩展。可以用可丢弃的本地数据库运行相同测试：

```shell
go build -o /tmp/pzero-goctl github.com/zeromicro/go-zero/tools/goctl
GOCTL=/tmp/pzero-goctl \
PZERO_TEST_POSTGRES_DSN='postgres://postgres:postgres@127.0.0.1:5432/pzero_test?sslmode=disable' \
go test ./cmd/pzero/internal/command/gen/genmodel -run TestPostgresGeneratedModels -count=1
```

测试创建独立 schema 并清理，所用数据库账号需有创建 schema 的权限。没有配置环境变量时本地测试跳过；配置后连接或初始化失败会使测试失败。

## 从 jzero 迁移注意

1. 保留 `desc/sql` snapshot；不要删，也不会再阻塞 `pzero gen`。
2. 配置 `model-datasource: true` 与 `model-datasource-url`，并把 **全部** 需要注册的表写入 `model-datasource-table`（`model.go` 会按该列表全量重写）。
3. 先 `migrate up`，再 `pzero gen`（无 `--desc`）或 `pzero gen model`，从数据库结构生成 model。
4. 生成 API 已 slim：`InsertV2` 改为 `Insert`；不再生成 `WithTable`、`FindSelectedColumnsByCondition`、按唯一索引的 `FindOneBy*`。业务侧请改用 `FindOneByCondition` 或 custom 方法。
5. 生成 import 会切到 `github.com/polpo-space/pzero/core/stores/{modelx,condition}`。
