# CRUD Operations

## Generated Methods

- `Insert`
- `InsertV2`
- `BulkInsert`
- `FindOne`
- `FindByCondition`
- `FindOneByCondition`
- `CountByCondition`
- `PageByCondition`
- `Update`
- `UpdateFieldsByCondition`
- `Delete`
- `DeleteByCondition`
- `WithTable(...).Method(...)`

## Guidance

- Use generated CRUD methods for normal operations
- Use `InsertV2` when you need the inserted ID
- Use `UpdateFieldsByCondition` for partial updates
- Use `condition.NewChain()` to build conditions
- Only write custom SQL for queries the generated methods cannot express cleanly
