# CRUD Operations

## Generated Methods

- `Insert`
- `BulkInsert`
- `FindOne`
- `FindOneForUpdate`
- `FindByCondition`
- `FindFieldsByCondition`
- `FindOneByCondition`
- `FindOneFieldsByCondition`
- `CountByCondition`
- `PageByCondition`
- `Update`
- `UpdateFieldsByCondition`
- `UpdateFieldsByConditionAffected`
- `Delete`
- `DeleteByCondition`

## Guidance

- Use `Insert` for single-row inserts on PostgreSQL
- Use `FindOneByCondition` with `condition` for unique-index lookups (per-index `FindOneBy*` helpers are not generated)
- Use `FindFieldsByCondition` instead of the removed `FindSelectedColumnsByCondition`
- Use `UpdateFieldsByCondition` for partial updates
- Use `condition.NewChain()` to build conditions
- Only write custom SQL for queries the generated methods cannot express cleanly

## Conditional Updates and Affected Rows

Use `UpdateFieldsByCondition` for the existing error-only contract. Use `UpdateFieldsByConditionAffected` when the caller needs `(int64, error)`, for example to compare and update a state. Include the row's primary key and expected state in the conditions; the method also supports bulk updates and adds no state precondition itself.

```go
affected, err := userModel.UpdateFieldsByConditionAffected(ctx, nil,
    map[string]any{string(usersmodel.Status): nextStatus},
    condition.NewChain().
        Equal(usersmodel.Id, userID).
        Equal(usersmodel.Status, expectedStatus).
        Build()...,
)
```

Interpret both return values:

- With nonempty update fields and no error, `affected == 1` means this row matched; `affected == 0` means the condition did not match. Map that conflict to the application's registered business error.
- Nil or empty update maps are deliberate no-ops returning `(0, nil)`. Validate that CAS update fields are nonempty before interpreting zero as a conflict.
- SQL may succeed and cache deletion may fail. A non-nil error can accompany a nonzero row count; preserve both and do not blindly retry or report a state conflict. If reading `RowsAffected` itself fails, the count is unavailable. Combined errors remain inspectable with `errors.Is`.
- When a transaction session is supplied, an affected row count does not prove commit; the application owns commit/rollback and its final result.

## Transaction Row Locks and Cache

Organize the transaction in the application. Only pass the original session from `SqlConn.TransactCtx` to `FindOneForUpdate` and subsequent writes:

```go
// Use a model generated without caching for this transaction path.
err := conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
    user, err := userModel.FindOneForUpdate(ctx, session, userID)
    if err != nil {
        return err
    }
    user.Status = nextStatus // Check application state rules before changing it.
    return userModel.Update(ctx, session, user)
})
```

`FindOneForUpdate` issues `SELECT ... FOR UPDATE` and always bypasses the cache. Nil and ordinary `sqlx.SqlConn` arguments return `modelx.ErrTransactionRequired`; arbitrary Session wrappers cannot be proven transactional. A missing row returns the model's `ErrNotFound`, including when the underlying error is wrapped. Locks last until the caller commits or rolls back; the method creates no implicit transaction.

Cached write methods currently delete cache entries **before transaction commit**. A concurrent ordinary read can refill the old value before commit. Bypassing the cache for the locking read does not fix this race, and the framework has no automatic after-commit invalidation. Generate the model without caching (for example, `gen.model-cache: false`) for a strongly consistent transaction path, or implement an application-owned after-commit invalidation and consistency strategy.

## Handwritten SQL Extensions

Use `condition.BuildSelect`, `BuildUpdate`, and `BuildDelete` to build PostgreSQL statements. `SelectByWhereRawSql` mutates the supplied builder by selecting PostgreSQL and appending predicates; supply a builder owned by the current query.
