# Database Connection

## Example Config

```yaml
sqlx:
  driverName: "mysql"
  dataSource: "root:password@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  host: "127.0.0.1:6379"
  type: "node"
  pass: "yourpassword"
```

## Guidance

- Initialize database and cache in `ServiceContext`
- Do not create connections inside handlers or logic
- Use cache only when read/write patterns justify it
