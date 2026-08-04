# Database Connection

## Example Config

```yaml
sqlx:
  driverName: "pgx"
  dataSource: "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"

redis:
  host: "127.0.0.1:6379"
  type: "node"
  pass: "yourpassword"
```

## Guidance

- Use PostgreSQL through the `pgx` driver in generated API and RPC services
- Keep the same `sqlx` configuration for models and service migration commands
- Initialize database and cache in `ServiceContext`
- Do not create connections inside handlers or logic
- Use cache only when read/write patterns justify it
