---
title: Add descriptor files
icon: /icons/proicons-tag-add.svg
order: 3.1
---

pzero generates code from descriptor files (desc) and config:

* desc/api: api descriptor language, generate http server/client code. [User guide](../guide/api.md)
* desc/proto: proto descriptor language, generate grpc server/client code. [User guide](../guide/proto.md)
* desc/sql: schema snapshot of the current structure; **not** used for model generation. [User guide](../guide/model.md)

Model code is generated from a remote datasource:

* model datasource: `model-datasource: true` + `model-datasource-url`. [User guide](../guide/model.md)

Schema changes belong in `desc/sql_migration/` via the service binary `migrate create` / `migrate up`.

## Add api file

Will add api file under desc/api folder

```shell
# group is test
pzero add api test
# group is test/test1
pzero add api test/test1
```

## Add proto file

Will add proto file under desc/proto folder

```shell
# Service is Test
pzero add proto test
# Service is TestTest1
pzero add proto test/test1
```

## Add sql snapshot

Creates a PostgreSQL schema snapshot placeholder under `desc/sql`.  
This is not model input; enable datasource and run `pzero gen` to generate models.

```shell
# table name is test
pzero add sql test
```
