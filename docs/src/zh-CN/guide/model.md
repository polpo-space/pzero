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
* model 生成只支持远程数据源模式
* `desc/sql` 下的 SQL DDL 文件不再作为 `pzero gen` 的 model 输入

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
```

生成的 `internal/model/model.go` 会注册所选表对应的模型代码。

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
```

生成的 `internal/model/model.go` 会与所选 PostgreSQL 数据源集合保持一致。

## 默认生成如下方法

* InsertV2: 插入单条数据（PostgreSQL 通过 `RETURNING` 回填自增主键）
* BulkInsert: 批量插入数据
* Update: 根据主键更新单条数据
* Delete: 根据主键删除单条数据
* FindOne: 根据主键查询单条数据
* FindByCondition: 条件查询
* FindFieldsByCondition: 条件查询(指定查询字段)
* FindOneByCondition: 条件查询单条数据（按唯一索引查询请用此方法，不再生成 `FindOneBy*` 系列）
* FindOneFieldsByCondition: 条件查询单条数据(指定查询字段)
* CountByCondition: 条件查询总数
* PageByCondition: 条件分页查询
* UpdateFieldsByCondition: 条件更新指定字段
* DeleteByCondition: 条件删除

具体使用请参阅: [condition组件](../component/condition.md)
