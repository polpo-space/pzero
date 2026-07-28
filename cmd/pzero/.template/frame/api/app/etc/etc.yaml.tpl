rest:
    name: {{ .APP }}-api
    host: 0.0.0.0
    port: 8001

log:
    serviceName: {{ .APP }}
    encoding: plain
    level: info
    mode: console
{{ if has "model" .Features }}
sqlx:
    driverName: "pgx"
    dataSource: "postgres://postgres:postgres@127.0.0.1:5432/{{ .APP }}?sslmode=disable"
{{ end }}{{ if has "redis" .Features }}
redis:
    host: "127.0.0.1:6379"
    type: "node"
    pass: "123456"{{ end }}