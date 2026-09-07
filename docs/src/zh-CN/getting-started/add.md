---
title: 新增可描述文件
icon: /icons/proicons-tag-add.svg
order: 3.1
---

pzero 根据可描述文件(desc)与配置生成代码:

* desc/api: api 可描述语言, 生成 http 服务端/客户端代码. [使用指南](../guide/api.md)
* desc/proto: proto 可描述语言, 生成 grpc 服务端/客户端代码. [使用指南](../guide/proto.md)
* desc/sql: schema snapshot（当前结构镜像），**不**用于 model 生成. [使用指南](../guide/model.md)

Model 代码通过远程数据源配置生成:

* model datasource: `model-datasource: true` + `model-datasource-url`. [使用指南](../guide/model.md)

Schema 变更请用 migration：`desc/sql_migration/`，通过服务二进制 `migrate create` / `migrate up`。

## 新增 api 文件

将在 desc/api 文件夹下新增 api 文件

```shell
# group 为 test
pzero add api test
# group 为 test/test1
pzero add api test/test1
```

## 新增 proto 文件

将在 desc/proto 文件夹下新增 proto 文件

```shell
# Service 为 Test
pzero add proto test
# Service 为 TestTest1
pzero add proto test/test1
```

## 新增 sql snapshot

将在 `desc/sql` 下新增 schema snapshot 占位文件（PostgreSQL DDL）。  
这不是 model 输入；生成 model 请配置 datasource 后执行 `pzero gen`。

```shell
# table 名为 test
pzero add sql test
```
