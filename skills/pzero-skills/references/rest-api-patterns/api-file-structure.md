# REST API Patterns

## Critical API File Rules

Every `.api` file must follow these rules:

1. Set `go_package`
2. Set `group` in `@server`
3. Set `compact_handler: true` in `@server`

Example:

```api
info(
    title: "User API"
    go_package: "user"
)

@server(
    prefix: /api/v1
    group: user
    compact_handler: true
)
service user-api {
    @handler Create
    post /users (CreateRequest) returns (CreateResponse)
}
```

## Core Architecture

pzero REST APIs follow a strict three-layer architecture:

1. `internal/handler/`: HTTP concerns only
2. `internal/logic/`: business logic
3. `internal/svc/`: dependency injection

## Guidance

- Keep handlers thin
- Put business rules in logic
- Wire dependencies in service context
- Avoid redundant prefixes when `group` is already set

## Business Errors

API projects ship `internal/errcode/errcode.go`. Every code is declared and registered in one line;
`internal/middleware/response.go` converts the error into `{code, msg, data}` via `status.FromError`.

Declare a new code:

```go
// internal/errcode/errcode.go
var (
    UserNotFound  = register(10001, "user not found")
    UserDisabled  = register(10002, "user disabled")
)
```

Return it from logic:

```go
import (
    "github.com/polpo-space/pzero/core/status"

    "example.com/app/internal/errcode"
)

user, err := l.svcCtx.Model.User.FindOne(l.ctx, req.Id)
if errors.Is(err, usermodel.ErrNotFound) {
    return nil, status.Error(errcode.UserNotFound)
}
if err != nil {
    return nil, status.Wrap(errcode.Internal, err)
}
if user.Disabled {
    return nil, status.ErrorMessage(errcode.UserDisabled, "user "+req.Id+" is disabled")
}
```

Rules:

- Group codes by module with a dedicated range (for example `10001-10999` for users)
- Never pass a raw `status.Code` literal to `status.Error`; unregistered codes degrade to `500`
- Use `status.Wrap` to keep the underlying error for logging while returning a stable code to callers
