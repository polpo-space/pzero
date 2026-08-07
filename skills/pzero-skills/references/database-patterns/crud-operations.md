# CRUD Operations

## Generated Methods

- `InsertV2`
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

- Use `InsertV2` for single-row inserts on PostgreSQL
- Use `FindOneByCondition` with `condition` for unique-index lookups (per-index `FindOneBy*` helpers are not generated)
- Use `FindFieldsByCondition` instead of the removed `FindSelectedColumnsByCondition`
- Use `UpdateFieldsByCondition` for partial updates
- Use `condition.NewChain()` to build conditions
- Only write custom SQL for queries the generated methods cannot express cleanly
