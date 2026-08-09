// Package job 提供由 gocron 驱动的定时任务调度器，实现 go-zero 的
// service.Service，可直接加入 service group 与 RPC/API 服务合并部署。
//
// 服务侧只需要提供配置和一组具名 handler，并发控制、单例、超时、panic
// 恢复、优雅退出都由本包统一处理。
package job

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"
	"time"

	"github.com/eddieowens/opts"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var _ service.Service = (*Server)(nil)

// Handler 是一次 job 执行。ctx 会在进程退出时被取消，
// 配置了 timeout 时还会带上超时，handler 应当响应取消。
type Handler func(context.Context) error

// NamedHandler 把 job 名与 handler 绑定，名字必须与配置里的 key 一致。
type NamedHandler struct {
	Name    string
	Handler Handler
}

// Named 构造一个 NamedHandler。
func Named(name string, handler Handler) NamedHandler {
	return NamedHandler{
		Name:    name,
		Handler: handler,
	}
}

// Server 是 job 调度器，实现 service.Service。
type Server struct {
	scheduler       gocron.Scheduler
	shutdownTimeout time.Duration
	registered      []string
}

// New 校验配置与 handler 并构建调度器。配置里的 job 与注册的 handler
// 必须严格一一对应，任何一边多出或缺失都会返回错误，避免改了 yaml
// 忘了写代码（或反过来）时任务静默不执行。
func New(cfg Config, handlers []NamedHandler, op ...opts.Opt[Options]) (*Server, error) {
	o := opts.DefaultApply(op...)

	cfg, location, err := cfg.normalize()
	if err != nil {
		return nil, err
	}
	if cfg.ShutdownTimeout > graceWindow {
		logx.Errorf("job: shutdownTimeout %v exceeds the %v go-zero allows before force kill, "+
			"raise zrpc.shutdown.waitTime as well", cfg.ShutdownTimeout, graceWindow)
	}

	registry, err := buildRegistry(cfg, handlers)
	if err != nil {
		return nil, err
	}

	scheduler, err := gocron.NewScheduler(schedulerOptions(cfg, location)...)
	if err != nil {
		return nil, fmt.Errorf("job: create scheduler: %w", err)
	}

	registered, err := addJobs(scheduler, cfg, registry, o.Namespace)
	if err != nil {
		_ = scheduler.Shutdown()
		return nil, err
	}

	return &Server{
		scheduler:       scheduler,
		shutdownTimeout: cfg.ShutdownTimeout,
		registered:      registered,
	}, nil
}

// Start 实现 service.Service，非阻塞。
func (s *Server) Start() {
	logx.Infof("job: starting scheduler, %d job(s): %v", len(s.registered), s.registered)
	s.scheduler.Start()
}

// Stop 实现 service.Service，取消所有运行中 job 的 ctx 并等待其结束。
func (s *Server) Stop() {
	logx.Info("job: stopping scheduler...")

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.scheduler.ShutdownWithContext(ctx); err != nil {
		logx.Errorf("job: shutdown scheduler: %v", err)
		return
	}
	logx.Info("job: scheduler stopped")
}

func schedulerOptions(cfg Config, location *time.Location) []gocron.SchedulerOption {
	options := []gocron.SchedulerOption{
		gocron.WithLocation(location),
		gocron.WithStopTimeout(cfg.ShutdownTimeout),
		gocron.WithLogger(schedulerLogger{}),
		gocron.WithGlobalJobOptions(
			gocron.WithEventListeners(gocron.AfterJobRunsWithError(onError)),
		),
	}

	if cfg.Concurrency.Limit > 0 {
		// normalize 已经校验过 mode。
		mode, _ := cfg.Concurrency.Mode.gocronMode()
		options = append(options, gocron.WithLimitConcurrentJobs(uint(cfg.Concurrency.Limit), mode))
	}

	return options
}

func addJobs(scheduler gocron.Scheduler, cfg Config, registry map[string]Handler, namespace string) ([]string, error) {
	registered := make([]string, 0, len(cfg.Jobs))

	for _, name := range sortedNames(cfg.Jobs) {
		spec, fullName := cfg.Jobs[name], qualify(namespace, name)
		if !spec.Enable {
			logx.Infof("job: %s disabled by config", fullName)
			continue
		}

		jobOptions := []gocron.JobOption{gocron.WithName(fullName)}
		// normalize 已经校验过 overlap。
		if mode, singleton, _ := spec.Overlap.singletonMode(); singleton {
			jobOptions = append(jobOptions, gocron.WithSingletonMode(mode))
		}

		handler, timeout := registry[name], spec.Timeout
		task := gocron.NewTask(func(ctx context.Context) error {
			return run(ctx, handler, timeout)
		})

		if _, err := scheduler.NewJob(spec.definition(), task, jobOptions...); err != nil {
			return nil, fmt.Errorf("job %q: %w", name, err)
		}

		registered = append(registered, fullName)
		logx.Infof("job: registered %s cron=%q every=%v overlap=%s timeout=%v",
			fullName, spec.Cron, spec.Every, spec.Overlap, spec.Timeout)
	}

	return registered, nil
}

// run 为单次执行派生 ctx 并兜住 panic。
//
// gocron 的 job ctx 在整个 job 生命周期内复用，所以每次执行的超时必须
// 在这里派生；panic 也必须在这里恢复——gocron 只有注册了 panic 监听器
// 才会 recover，而那时栈已经展开，拿不到 panic 现场。
func run(ctx context.Context, handler Handler, timeout time.Duration) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%w: %v\n%s", ErrPanic, recovered, debug.Stack())
		}
	}()

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return handler(ctx)
}

// buildRegistry 校验配置与 handler 严格一一对应。
func buildRegistry(cfg Config, handlers []NamedHandler) (map[string]Handler, error) {
	registry := make(map[string]Handler, len(handlers))

	for _, h := range handlers {
		switch {
		case h.Name == "":
			return nil, ErrEmptyHandlerName
		case h.Handler == nil:
			return nil, fmt.Errorf("%w: %s", ErrNilHandler, h.Name)
		}
		if _, ok := registry[h.Name]; ok {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateHandler, h.Name)
		}
		if _, ok := cfg.Jobs[h.Name]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrJobNotConfigured, h.Name)
		}
		registry[h.Name] = h.Handler
	}

	for _, name := range sortedNames(cfg.Jobs) {
		if _, ok := registry[name]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrHandlerNotFound, name)
		}
	}

	return registry, nil
}

func sortedNames(jobs map[string]Spec) []string {
	names := make([]string, 0, len(jobs))
	for name := range jobs {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

func qualify(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return namespace + "." + name
}

func onError(_ uuid.UUID, name string, err error) {
	logx.Errorf("job: %s failed: %v", name, err)
}

// schedulerLogger 把 gocron 的内部日志接到 logx。
type schedulerLogger struct{}

func (schedulerLogger) Debug(msg string, args ...any) { logx.Debugf("%s %v", msg, args) }

func (schedulerLogger) Info(msg string, args ...any) { logx.Infof("%s %v", msg, args) }

func (schedulerLogger) Warn(msg string, args ...any) { logx.Errorf("%s %v", msg, args) }

func (schedulerLogger) Error(msg string, args ...any) { logx.Errorf("%s %v", msg, args) }
