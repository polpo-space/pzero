---
title: 生成服务端代码
icon: /icons/vscode-icons-folder-type-api-opened.svg
order: 4
---

pzero 生成代码命令极其精简, 仅需 `pzero gen` 就能自动识别所有的可描述文件/配置, 完成代码的生成.

通过上一篇文档的 `pzero add` 命令添加可描述文件后, 执行 `pzero gen` 即可看到生成的文件了.

## 生成代码

::: code-tabs#shell

@tab pzero

```bash
cd your_project
pzero gen
```

@tab Docker

```bash
cd your_project
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest gen
```
:::

Model 生成需要在配置中开启 `model-datasource: true` 并提供 `model-datasource-url`。`desc/sql` 仅为 schema snapshot，不会触发 model 生成。

## 只生成 model

`pzero gen model` 只跑 model 段，即使目录里已有 `desc/api` 或 proto，也不会重写 api/rpc。

该命令本身即开启 datasource 模式。URL 可通过 `--model-datasource-url` 或 `.pzero.yaml` 提供。

```shell
pzero gen model
pzero gen model --model-datasource-url "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"
```

## 指定 desc 生成代码

`--desc` 用于限定 **api/proto** 生成范围；指定后会跳过 model 生成。

```shell
pzero gen --desc desc/api/xx.api
pzero gen --desc desc/proto/xx.proto
```

## 忽略指定 desc 生成代码

```shell
pzero gen --desc-ignore desc/api/xx.api
pzero gen --desc-ignore desc/proto/xx.proto
```

更多用法请参阅: [pzero 指南](../guide/jzero.md)

## 退役功能迁移

- 删除 `--git-change`、`gen.git-change` 和 `PZERO_GEN_GIT_CHANGE`。`pzero gen` 完整生成当前应用的描述文件；CI 在应用层选择范围，仍可用 `--desc` / `--desc-ignore` 明确限定输入。`pzero format --git-change` 继续支持。
- 删除 `--route2code`、`gen.route2code`、`gen.swagger.route2code` 及对应环境变量。Pzero 不再生成权限编码映射，也不再将其注入 Swagger 描述。已有使用方应将权限映射移入应用维护的代码、更新调用方，再删除旧生成文件 `internal/handler/route2code.go`。
- 生成器不再写 IDE 跳转元数据。读取 `~/.pzero/desc-metadata` 的 GoLand 插件需要更换跳转数据来源。已有元数据不会自动删除，可手动清理。
- Model 每次只接受一个 PostgreSQL URL，可写为字符串或单元素 YAML 列表。多个 URL 和 `database.table` 表名会在写模型文件前报错。使用不带前缀的表名，通过 `model-schema` 指定 PostgreSQL schema。不同数据库应在独立项目/输出目录生成；在同一项目依次生成会覆盖 `model.go` 注册。
- 自定义 `model/model.go.tpl` 应使用 `.Imports`（字符串列表）与 `.TableInfos`。不再提供 `.ImportsWithAlias`、`.MutiModels`、`.MutiModelsWithAlias` 及表的 `Alias`/`FullName`/`WithCache` 字段。缓存过期配置和单表构造函数继续支持。

退役的命令行开关会报 unknown flag；退役的配置和环境变量会给出迁移错误，避免静默改变行为。
