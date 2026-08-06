# RPC Job Patterns

Job（定时任务）可与 gRPC 合并部署在同一进程，通过 `pzero new --features job` 生成骨架。

## Scope

- Supported: merge deploy with `service.ServiceGroup`, config-driven schedules
- Not generated: standalone `cmd/job`, job desc files for `pzero gen`, distributed locks / etcd election

## Layout

```text
internal/
  config/config.go          # JobConf + JobSpec
  job/example_job.go        # thin handler: func(ctx) error
  logic/job/example_logic.go
  server/job.go             # wire name -> handler, register from config
cmd/server.go               # if c.Job.Enable { group.Add(jobServer) }
etc/etc.yaml                # job.enable / workers / timezone / jobs
```

## Config

Schedules live in yaml — do not hardcode cron in `AddFunc`.

```yaml
job:
  enable: true
  workers: 5
  timezone: Asia/Shanghai
  jobs:
    orderTimeout:
      enable: true
      cron: "*/30 * * * * *"
    paymentExpire:
      enable: true
      cron: "0 */1 * * * *"
```

Scaffold defaults:

```yaml
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
```

- `job.enable`: process-level switch (whether JobServer joins ServiceGroup)
- `job.workers`: max concurrent job executions (semaphore)
- `job.timezone`: cron location (`cron.WithLocation`)
- `job.jobs.<name>`: per-job enable + cron (6-field with seconds, or `@every`)

## Handler -> Logic

Handler only creates logic and calls it. Handler map keys **must match** `job.jobs` names. Startup **fail-fast**:

- every key under `job.jobs` must have a registered handler (even if `enable: false`)
- enabled jobs must have a non-empty `cron`

Missing handler or empty cron aborts process start (`logx.Must`), so yaml/config drift cannot silently ship.

```go
handlers := map[string]func(context.Context) error{
	"exampleInterval": exampleJob.ExampleInterval,
	"orderTimeout":    orderJob.OrderTimeout,
}
```

## Adding a new job

1. Add handler method under `internal/job/`
2. Implement logic under `internal/logic/job/`
3. Register `name -> handler` in `internal/server/job.go`
4. Add `job.jobs.<name>` with `enable` + `cron` in `etc/etc.yaml`

Changing schedule / toggling a job should only need yaml (and redeploy), not code edits to cron strings.

## Lifecycle

`JobServer.Start` uses `cron.Start()` (async). `Stop` waits on `cron.Stop()` until running jobs finish.

## Multi-instance note

Merged jobs run on every replica that has `job.enable: true`. If only one instance should run a task, implement your own distributed lock (e.g. Redis) in logic — pzero does not generate lock helpers.
