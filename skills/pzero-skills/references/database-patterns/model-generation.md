# Model Generation

## Overview

pzero generates models from a PostgreSQL remote datasource only.

`desc/sql` is a schema snapshot for review/alignment. It may coexist with datasource mode and is never model gen input.

## Datasource Mode

```yaml
gen:
  model-driver: postgres
  model-datasource: true
  model-datasource-url:
    - "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"
  model-datasource-table:
    - users
```

```bash
# migrate first, then generate models from the live schema
pzero gen
```

`--desc` scopes api/proto only and skips model generation. Do not pass `.sql` to `--desc`.

## Common Methods

- `Insert`
- `BulkInsert`
- `FindOne`
- `FindByCondition`
- `FindFieldsByCondition`
- `FindOneByCondition`
- `FindOneFieldsByCondition`
- `CountByCondition`
- `PageByCondition`
- `Update`
- `UpdateFieldsByCondition`
- `Delete`
- `DeleteByCondition`

## Guidance

- Regenerate models after applying migrations
- Keep `model-datasource-table` complete: `internal/model/model.go` is fully rewritten from that list
- Pair generation with migration discipline (`desc/sql_migration/`)
- Keep `desc/sql` snapshots in sync for review, not for gen
