---
title: 数据库迁移
icon: /icons/carbon-migrate.svg
star: true
order: 5.5
---

* 使用 `model` feature 生成的 API/RPC 项目会自带 `migrate` 子命令
* migration 从 `desc/sql_migration` 读取，并通过生成的服务二进制显式执行
* migration 执行只支持 PostgreSQL 的 `pgx` driver
* 启动服务不会自动执行 migration
* 参考 [最佳实践](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md) 编写 migration 文件

## 配置

```yaml
sqlx:
  driverName: "pgx"
  dataSource: "postgres://postgres:postgres@127.0.0.1:5432/jzero-admin?sslmode=disable"
```

## 创建

```shell
go run . migrate create add_users
```

## 升级与回滚

```shell
# 应用所有待执行的 migrations
go run . migrate up --config etc/etc.yaml
# 最多应用三个 migrations
go run . migrate up --steps 3 --config etc/etc.yaml
# 回滚一个 migration
go run . migrate down --steps 1 --config etc/etc.yaml
```

## 版本控制

```shell
go run . migrate version --config etc/etc.yaml
go run . migrate goto <version> --config etc/etc.yaml
# 仅在人工确认数据库状态后清理 dirty 状态
go run . migrate force <version> --config etc/etc.yaml
```
