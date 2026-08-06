# RPC Job Patterns

Job（定时任务）可与 gRPC 合并部署在同一进程，通过 `pzero new --features job` 生成骨架。

## Scope

- Supported: merge deploy with `service.ServiceGroup`
- Not generated: standalone `cmd/job`, job desc files for `pzero gen`, distributed locks / etcd election

## Layout

```text
internal/
  config/config.go          # Job.Enable
  job/example_job.go        # thin handler: func(ctx) error
  logic/job/example_logic.go
  server/job.go             # robfig/cron + service.Service
cmd/server.go               # if c.Job.Enable { group.Add(jobServer) }
etc/etc.yaml                # job.enable: false
```

## Enable

```yaml
job:
  enable: true
```

When enabled, `cmd/server` adds `JobServer` to the same `ServiceGroup` as zrpc.

## Handler -> Logic

Handler only creates logic and calls it:

```go
func (j *ExampleJob) EveryMinute(ctx context.Context) error {
	return joblogic.NewExampleLogic(ctx, j.svcCtx).EveryMinute()
}
```

Put business code in `internal/logic/job/`.

## Register jobs

Jobs are registered in code inside `internal/server/job.go` (not YAML):

```go
c := cron.New(cron.WithSeconds())

// fixed interval
_, _ = c.AddFunc("@every 5s", func() { ... })

// 6-field cron (seconds first)
_, _ = c.AddFunc("0 * * * * *", func() { ... })
```

`JobServer.Start` uses `cron.Run()` (blocking). `Stop` uses `cron.Stop()`.

## Adding a new job

1. Add a method on a job handler under `internal/job/`
2. Implement logic under `internal/logic/job/`
3. Register it in `internal/server/job.go` with `AddFunc`
4. Keep `job.enable` as the process-level switch

## Multi-instance note

Merged jobs run on every replica that has `job.enable: true`. If only one instance should run a task, implement your own distributed lock (e.g. Redis) in logic — pzero does not generate lock helpers.
