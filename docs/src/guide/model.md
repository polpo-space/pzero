---
title: model guide
icon: /icons/uil-database.svg
star: true
order: 4
---

## Introduction

pzero generates database code into `internal/model` from a PostgreSQL datasource.

For easier usage, pzero automatically generates `internal/model/model.go` to register all generated models.

The model generator is PostgreSQL-only:
* `gen.model-driver` accepts `postgres` (recommended) or `pgx`
* model generation only supports datasource input
* SQL DDL input under `desc/sql` is no longer used by `pzero gen` for model generation

## Features

* Supports multi-datasource code generation
* Supports redis/custom cache
* Keeps generated model registration consistent across single-schema and multi-schema projects

## Cache Configuration

pzero supports flexible cache configuration for generated models:

* **model-cache**: Enable or disable cache for generated models (default: false)
* **model-cache-table**: Specify which tables to cache (default: * for all tables)
* **model-cache-expiry-table**: Configure custom cache expiry times for specific tables

### Cache Expiry Configuration

The `model-cache-expiry-table` option allows you to set different cache expiry times for different tables:

```yaml
gen:
  model-cache-expiry-table:
    - table: manage_user
      expiry: 3600              # cache expiry for found data (in seconds)
      not-found-expiry: 60      # cache expiry for not-found data (in seconds)
    - table: manage_role
      expiry: 7200
      not-found-expiry: 120
```

**Fields:**
* **table**: The table name to apply the cache expiry settings
* **expiry**: Cache duration in seconds for successfully queried data (default: system default)
* **not-found-expiry**: Cache duration in seconds for not-found queries (default: system default)

This configuration is particularly useful for:
* Tables with different data update frequencies
* Reducing cache misses for frequently accessed reference data
* Optimizing cache hit rates for specific business scenarios

### NewOriginal Functions

The `model-new-original` option controls whether to generate `NewOriginal` functions for each table:

```yaml
gen:
  model-new-original: true  # default: false
```

When enabled, pzero generates additional `NewOriginal*XxxModel` functions alongside the standard `NewModel` function. These functions provide:

* **Direct model initialization**: Create individual model instances without initializing all models
* **Flexible cache configuration**: Each model can be initialized with custom cache options
* **Better performance**: Only initialize the models you actually need

**Example generated code:**

```go
// Standard initialization (all models)
func NewModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) Model {
    return Model{
        ManageUser: manage_user.NewManageUserModel(conn, op...),
        // ... other models
    }
}

// Individual model initialization (when model-new-original: true)
func NewOriginalManageUserModel(conn sqlx.SqlConn, c cache.CacheConf, op ...cache.Option) manage_user.ManageUserModel {
    return manage_user.NewOriginalManageUserModel(conn, c, op...)
}
```

**Usage scenarios:**
* Microservices where only specific models are needed
* Applications requiring different cache configurations per model
* Performance optimization by avoiding unnecessary model initialization

## Generate code based on remote datasource

```yaml
gen:
  model-driver: postgres
  # whether to generate database code with cache
  model-cache: true
  # cache tables, default is *(all)
  model-cache-table:
    - manage_user
  # set cache expiry for specific tables (in seconds)
  model-cache-expiry-table:
    - table: manage_user
      expiry: 3600
      not-found-expiry: 60
  # generate NewOriginal functions for each table (default: false)
  model-new-original: true
  # postgres model generation requires datasource mode
  model-datasource: true
  # postgres datasource configuration
  model-datasource-url: "postgres://postgres:postgres@127.0.0.1:5432/jzero-admin?sslmode=disable"
  # schema, default is public
  model-schema: public
  # Ignore columns while creating or updating rows, default is create_at,created_at,create_time,update_at,updated_at,update_time
  model-ignore-columns: ["create_time", "update_time"]
  # which tables to use, default is *(all)
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

Generated `internal/model/model.go` contains the registered models for the selected tables.

pzero supports multi-datasource model generation. Use `database.table` in `model-datasource-table` to map a table to a specific datasource URL.

For example, add `jzero-admin_log.operate_log`:

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

Generated `internal/model/model.go` stays consistent with the selected PostgreSQL datasource set.

## Default generated methods

* InsertV2: Insert single row (PostgreSQL `RETURNING` populates auto-increment primary key)
* BulkInsert: Batch insert
* Update: Update single row by primary key
* Delete: Delete single row by primary key
* FindOne: Query single row by primary key
* FindByCondition: Conditional query
* FindFieldsByCondition: Conditional query (specify query fields)
* FindOneByCondition: Conditional single row query (use this instead of per-index `FindOneBy*` helpers)
* FindOneFieldsByCondition: Conditional single row query (specify query fields)
* CountByCondition: Conditional total count
* PageByCondition: Conditional pagination query
* UpdateFieldsByCondition: Conditional update specified fields
* DeleteByCondition: Conditional delete

For detailed usage, see: [condition component](../component/condition.md)
