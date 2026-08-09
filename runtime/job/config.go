package job

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// LimitMode 描述调度器整体并发达到上限时的行为。
type LimitMode string

const (
	// LimitModeSkip 丢弃超出并发上限的这次触发。
	LimitModeSkip LimitMode = "skip"

	// LimitModeWait 排队等待空闲槽位。
	LimitModeWait LimitMode = "wait"
)

// OverlapPolicy 描述同一个 job 上一次还没跑完时的行为。
type OverlapPolicy string

const (
	// OverlapAllow 允许同一个 job 并发执行。
	OverlapAllow OverlapPolicy = "allow"

	// OverlapSkip 丢弃本次触发，等下一个周期。
	OverlapSkip OverlapPolicy = "skip"

	// OverlapWait 排队等上一次执行结束。
	OverlapWait OverlapPolicy = "wait"
)

// defaultShutdownTimeout 取 3s，因为 go-zero 收到退出信号后默认只留
// waitTime(5.5s) - wrapUpTime(1s) = 4.5s 就会强杀进程。
const defaultShutdownTimeout = 3 * time.Second

// graceWindow 是 go-zero 默认留给 Stop 的时间窗口，超过它的
// ShutdownTimeout 不可能生效，启动时给出告警。
const graceWindow = 4500 * time.Millisecond

// Config 是 job 调度器的配置，直接对应服务 yaml 中的 job 段。
type Config struct {
	// Enable 决定 JobServer 是否加入 service group，由调用方判断。
	Enable bool `json:",optional"`

	// Timezone 为 cron 表达式的时区，留空表示 time.Local。
	Timezone string `json:",optional"`

	// ShutdownTimeout 是等待运行中 job 结束的时间，留空取 3s。
	ShutdownTimeout time.Duration `json:",optional"`

	// Concurrency 是调度器级别的全局并发限制。
	Concurrency ConcurrencyConf `json:",optional"`

	// Jobs 以 job 名为 key，必须与注册的 handler 严格一一对应。
	Jobs map[string]Spec `json:",optional"`
}

// ConcurrencyConf 是调度器级别的并发限制。
type ConcurrencyConf struct {
	// Limit 为 0 表示不限制，防重叠请优先用 Spec.Overlap。
	Limit int `json:",optional"`

	// Mode 只在 Limit > 0 时有意义，留空取 wait。
	Mode LimitMode `json:",optional"`
}

// Spec 是单个 job 的配置。Cron 与 Every 二选一。
type Spec struct {
	Enable bool `json:",default=true"`

	// Cron 为 5 位或 6 位（含秒）表达式，也支持 @every 5s 这类描述符。
	Cron string `json:",optional"`

	// Every 表示固定间隔调度。
	Every time.Duration `json:",optional"`

	// Overlap 为上次未结束时的行为，留空取 skip。
	Overlap OverlapPolicy `json:",optional"`

	// Timeout 为单次执行的超时，留空表示不限制。超时只会取消 ctx，
	// handler 必须自己响应取消才能真正停下来。
	Timeout time.Duration `json:",optional"`
}

// normalize 填充零值并校验配置，返回规整后的副本与解析出的时区。
// 无论 job 是否 enable 都会校验，避免开关一打开才发现配置有问题。
func (c Config) normalize() (Config, *time.Location, error) {
	location := time.Local
	if c.Timezone != "" {
		loc, err := time.LoadLocation(c.Timezone)
		if err != nil {
			return Config{}, nil, fmt.Errorf("job: load timezone %q: %w", c.Timezone, err)
		}
		location = loc
	}

	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = defaultShutdownTimeout
	}

	if c.Concurrency.Limit < 0 {
		return Config{}, nil, ErrInvalidConcurrencyLimit
	}
	if c.Concurrency.Mode == "" {
		c.Concurrency.Mode = LimitModeWait
	}
	if _, err := c.Concurrency.Mode.gocronMode(); err != nil {
		return Config{}, nil, err
	}

	jobs := make(map[string]Spec, len(c.Jobs))
	for name, spec := range c.Jobs {
		normalized, err := spec.normalize()
		if err != nil {
			return Config{}, nil, fmt.Errorf("job %q: %w", name, err)
		}
		jobs[name] = normalized
	}
	c.Jobs = jobs

	return c, location, nil
}

func (s Spec) normalize() (Spec, error) {
	switch {
	case s.Cron != "" && s.Every != 0:
		return Spec{}, ErrScheduleConflict
	case s.Cron == "" && s.Every == 0:
		return Spec{}, ErrScheduleRequired
	case s.Every < 0:
		return Spec{}, ErrInvalidEvery
	}

	if s.Timeout < 0 {
		return Spec{}, ErrInvalidTimeout
	}

	if s.Overlap == "" {
		s.Overlap = OverlapSkip
	}
	if _, _, err := s.Overlap.singletonMode(); err != nil {
		return Spec{}, err
	}

	return s, nil
}

// definition 把 Spec 翻译成 gocron 的调度定义。
func (s Spec) definition() gocron.JobDefinition {
	if s.Cron != "" {
		return gocron.CronJob(s.Cron, true)
	}
	return gocron.DurationJob(s.Every)
}

func (m LimitMode) gocronMode() (gocron.LimitMode, error) {
	switch m {
	case LimitModeSkip:
		return gocron.LimitModeReschedule, nil
	case LimitModeWait:
		return gocron.LimitModeWait, nil
	default:
		return 0, fmt.Errorf("%w, got %q", ErrInvalidLimitMode, string(m))
	}
}

// singletonMode 返回 gocron 的 singleton 模式；第二个返回值为 false
// 表示不启用 singleton，即允许重叠执行。
func (p OverlapPolicy) singletonMode() (gocron.LimitMode, bool, error) {
	switch p {
	case OverlapAllow:
		return 0, false, nil
	case OverlapSkip:
		return gocron.LimitModeReschedule, true, nil
	case OverlapWait:
		return gocron.LimitModeWait, true, nil
	default:
		return 0, false, fmt.Errorf("%w, got %q", ErrInvalidOverlap, string(p))
	}
}
