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

- Set `sqlx.driverName: pgx` explicitly (`DriverName: pgx` with Go-style YAML keys). go-zero's config loader fills an omitted driver with `mysql`, not an empty string; `modelx.MustNewConn` rejects it with a configuration hint. The empty-string default applies only to directly constructed Go structs.
- `modelx.MustNewConn` supports PostgreSQL only and normalizes the `postgres` alias to `pgx`. Use `pgx` in shared service/migration config; the migrator requires that driver.
- The ORM uses PostgreSQL SQL without changing `sqlbuilder.DefaultFlavor`. Configure connections per instance; there is no runtime dialect switch.
- Keep the same `sqlx` configuration for models and service migration commands
- Initialize database and cache in `ServiceContext`
- Do not create connections inside handlers or logic
- Use cache only when read/write patterns justify it. Before combining cached writes with transactions, read [transaction and cache semantics](crud-operations.md#transaction-row-locks-and-cache).
