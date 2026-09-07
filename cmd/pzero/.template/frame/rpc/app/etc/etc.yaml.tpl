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
    timezone: Asia/Shanghai
    # 等待运行中任务结束的时间；应小于宿主服务实际留给 Stop 的退出窗口。
    shutdownTimeout: 3s
    # 调度器全局并发上限，0 表示不限制。
    # 防止同一个任务重叠请用 jobs.<name>.overlap，不要靠这里。
    concurrency:
        limit: 0
        mode: wait
    # key 必须与 internal/job/registry.go 里注册的 handler 严格一一对应。
    # cron 与 every 二选一；cron 支持 5 位、6 位（含秒）和 @every 5s 这类描述符。
    # overlap: wait 不能与 concurrency.limit > 0 同时使用。
    jobs:
        exampleInterval:
            enable: true
            every: 5s
            overlap: skip
        exampleMinute:
            enable: true
            cron: "0 * * * * *"
            overlap: skip
            timeout: 30s
{{ end }}
