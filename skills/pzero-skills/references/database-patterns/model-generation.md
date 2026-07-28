# Model Generation

## Overview

pzero can generate models from local SQL files or remote datasources.

## Local SQL Mode

```bash
pzero gen --desc desc/sql/users.sql
```

## Remote Datasource Mode

```bash
pzero gen
```

## Common Methods

- `Insert`
- `InsertV2`
- `BulkInsert`
- `FindOne`
- `FindByCondition`
- `PageByCondition`
- `Update`
- `UpdateFieldsByCondition`
- `Delete`
- `DeleteByCondition`

## Guidance

- Regenerate models after schema changes
- Pair generation with migration discipline
- Use local SQL mode for code-first workflows
