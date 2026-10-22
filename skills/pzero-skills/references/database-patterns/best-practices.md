# Database Best Practices

## Critical Rules

### Always use `condition.NewChain()`

```go
chain := condition.NewChain().
    Equal(usersmodel.Status, "active")
```

### Always import models with aliases

```go
import usersmodel "github.com/yourproject/internal/model/users"
```

### Always use `errors.Is()` from `github.com/pkg/errors`

```go
if errors.Is(err, usermodel.ErrNotFound) {
    return nil, errors.New("user not found")
}
```

### Choose the partial-update contract

Use `UpdateFieldsByCondition` when only the error is needed. For conditional state transitions, use `UpdateFieldsByConditionAffected` and follow [affected-row semantics](crud-operations.md#conditional-updates-and-affected-rows), including empty updates and cache failures.

```go
updateData := map[string]any{
    string(usersmodel.Name): req.Name,
}
```

## Guidance

- Use generated field constants
- Do not add redundant nil checks after `FindOne`
- Create migrations for every schema change
- Use `Insert` when you need inserted IDs (PostgreSQL `RETURNING`)
- For transactions or row locking, read [Transaction Row Locks and Cache](crud-operations.md#transaction-row-locks-and-cache) before choosing a cached model
