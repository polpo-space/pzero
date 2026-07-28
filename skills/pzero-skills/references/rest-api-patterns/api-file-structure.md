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
