zrpc:
    listenOn: 0.0.0.0:8000
    mode: dev
    name: {{ .APP }}

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
    pass: "123456"{{ end }}{{ if has "job" .Features }}
job:
    enable: false
    workers: 1
    timezone: Asia/Shanghai
    jobs:
        exampleInterval:
            enable: true
            cron: "@every 5s"
        exampleMinute:
            enable: true
            cron: "0 * * * * *"
{{ end }}
