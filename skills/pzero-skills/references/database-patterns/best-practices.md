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

### Use `UpdateFieldsByCondition` for partial updates

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
