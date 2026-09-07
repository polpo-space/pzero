# RPC Job Patterns

Job（定时任务）可与 gRPC 合并部署在同一进程，通过 `pzero new --features job` 生成骨架。

调度器本身在 `github.com/polpo-space/pzero/runtime/job`，底层是 `go-co-op/gocron/v2`。
服务侧只提供 **config + handlers**，不接触 gocron 类型。

## Scope

- Supported: merge deploy with `service.ServiceGroup`, config-driven schedules, per-job singleton, per-run timeout, panic recovery, graceful shutdown
- Not generated: standalone `cmd/job` binary, job desc files for `pzero gen`
- Not yet supported: distributed locker / leader election（多副本见下文）

## Layout

```text
internal/
  config/config.go          # Job runtimejob.Config
  job/registry.go           # Registry(svcCtx) -> []runtimejob.NamedHandler
  job/example_job.go        # thin handler: func(ctx) error
  logic/job/example_logic.go
  server/job.go             # runtimejob.New(config, registry, namespace)
cmd/server.go               # if c.Job.Enable { group.Add(jobServer) }
etc/etc.yaml                # job.*
```

## Config

Schedules live in yaml — never hardcode cron in Go.

```yaml
job:
  enable: true
  timezone: Asia/Shanghai
  shutdownTimeout: 3s
  concurrency:
    limit: 0
    mode: wait
  jobs:
    orderTimeout:
      enable: true
      every: 30s
      overlap: skip
    paymentReconcile:
      enable: true
      cron: "0 */5 * * * *"
      overlap: skip
      timeout: 4m
```

| 字段 | 说明 |
| --- | --- |
| `job.enable` | 进程级开关，决定 JobServer 是否加入 ServiceGroup |
| `job.timezone` | cron 时区，留空取 `time.Local` |
| `job.shutdownTimeout` | 等待运行中任务结束的时间，默认 3s |
| `job.concurrency.limit` | 调度器全局并发上限，`0` 表示不限制 |
| `job.concurrency.mode` | 达到上限时 `wait` 排队 / `skip` 丢弃，仅在 limit > 0 时有意义 |
| `job.jobs.<name>.enable` | 单个任务开关，默认 `true` |
| `job.jobs.<name>.cron` | 5 位、6 位（含秒）或 `@every 5s` 描述符 |
| `job.jobs.<name>.every` | 固定间隔，与 `cron` 互斥 |
| `job.jobs.<name>.overlap` | 上次未结束时 `skip` 跳过 / `wait` 排队 / `allow` 允许并发，默认 `skip` |
| `job.jobs.<name>.timeout` | 单次执行超时，留空不限制 |

`overlap: wait` 不能与 `concurrency.limit > 0` 同时使用。gocron 的全局
limiter 会先于 singleton limiter 接管任务，这个组合无法保证 per-job wait
队列语义，因此 pzero 会在启动时直接拒绝该配置。

### 防重叠用 overlap，不要用 concurrency.limit

`concurrency.limit` 是调度器**全局**闸门，limit=1 会让任务之间互相阻塞或互相丢弃。
防止同一个任务自己重叠，永远用 `overlap: skip`——这是默认值，对账、扫单、
拆账、backfill、清理这类任务都应该保持默认。

### shutdownTimeout 必须与宿主服务退出窗口对齐

go-zero 收到退出信号后：sleep `wrapUpTime`(默认 1s) → 通知 shutdown listener →
sleep `waitTime - wrapUpTime`(默认 4.5s) → 强杀进程。所以 `shutdownTimeout`
超过 4.5s 不会生效，必须同时调大：

```yaml
zrpc:
  shutdown:
    wrapUpTime: 1s
    waitTime: 40s
job:
  shutdownTimeout: 30s
```

`runtime/job` 不知道宿主进程实际配置的 shutdown window，因此不会拿 go-zero
默认值做固定阈值判断；生成服务或部署配置必须确保两者匹配。

## Handler -> Logic

Handler 只负责创建 logic 并调用。`Registry` 里的名字与 `job.jobs` 的 key
**必须严格一一对应**：

```go
func Registry(svcCtx *svc.ServiceContext) []runtimejob.NamedHandler {
	example := NewExampleJob(svcCtx)
	order := NewOrderJob(svcCtx)

	return []runtimejob.NamedHandler{
		runtimejob.Named("exampleInterval", example.ExampleInterval),
		runtimejob.Named("orderTimeout", order.OrderTimeout),
	}
}
```

以下情况一律启动失败，不会静默跳过：

- 配置里有 job，`Registry` 里没有对应 handler
- `Registry` 里有 handler，配置里没有对应 job
- handler 重名或为空
- cron / every 同时配置，或都没配置（**`enable: false` 的 job 同样校验**）
- cron 表达式非法、timezone 非法

因此多环境 yaml 必须列全所有 job，不需要的写 `enable: false`，不能直接删掉。

## Context

handler 拿到的 `ctx` 会在两种情况下被取消：进程退出（`JobServer.Stop`）、
配置了 `timeout` 且超时。DB / RPC / HTTP 调用都应该把这个 ctx 透传下去。

进程退出导致的 `context.Canceled` 视为正常 lifecycle 事件，记录为 Info；
job 自身 timeout 返回的 `context.DeadlineExceeded` 仍按执行失败记录为 Error。

注意 `timeout` 只取消 ctx，不会强杀 goroutine。handler 如果不响应取消，
会一直占着 singleton 槽位。

## Panic

handler panic 会被 `runtime/job` 捕获，转成带调用栈的 error 记进日志，
不会带走进程。**不要**在 handler 里自己写 recover 吞掉错误。

## Adding a new job

1. 在 `internal/job/` 加 handler 方法
2. 在 `internal/logic/job/` 实现逻辑
3. 在 `internal/job/registry.go` 的 `Registry` 里加一行 `runtimejob.Named(...)`
4. 在 `etc/etc.yaml` 的 `job.jobs` 下加同名配置

调整调度周期或开关任务只改 yaml，不用动代码。

## Lifecycle

`Start` 非阻塞（由 ServiceGroup 里的 RPC server 阻塞住进程）。
`Stop` 先取消所有 job 的 ctx，再等待运行中的任务结束，最长等 `shutdownTimeout`。

## Multi-instance note

合并部署时每个 `job.enable: true` 的副本都会独立调度，一次触发会执行 N 遍。
分布式锁 / leader 选举尚未支持，多副本场景下要么只在单实例 job runner 上开启，
要么业务自己保证幂等或抢锁。

## Job 命名与分布式锁

`runtimejob.WithNamespace(svcCtx.Config.Zrpc.Name)` 会把 job 名变成
`<service>.<job>`。这不是为了日志好看——将来接分布式锁时锁 key 就是 job 名，
没有服务前缀的话不同服务里的同名 job（比如都叫 `cleanup`）会互相抢锁。
