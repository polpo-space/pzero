package job

import "errors"

var (
	// ErrScheduleRequired 表示 job 既没有配置 cron 也没有配置 every。
	ErrScheduleRequired = errors.New("job: cron or every is required")

	// ErrScheduleConflict 表示 job 同时配置了 cron 和 every。
	ErrScheduleConflict = errors.New("job: cron and every are mutually exclusive")

	// ErrInvalidEvery 表示 every 不是正数。
	ErrInvalidEvery = errors.New("job: every must be positive")

	// ErrInvalidTimeout 表示 timeout 为负数。
	ErrInvalidTimeout = errors.New("job: timeout must not be negative")

	// ErrInvalidOverlap 表示 overlap 取值不在 allow/skip/wait 之内。
	ErrInvalidOverlap = errors.New("job: overlap must be one of allow, skip, wait")

	// ErrInvalidLimitMode 表示并发 mode 取值不在 skip/wait 之内。
	ErrInvalidLimitMode = errors.New("job: concurrency mode must be one of skip, wait")

	// ErrInvalidConcurrencyLimit 表示并发上限为负数。
	ErrInvalidConcurrencyLimit = errors.New("job: concurrency limit must not be negative")

	// ErrOverlapWaitWithConcurrency 表示全局并发限制与 job wait 队列组合使用。
	// gocron 的全局 limiter 优先于 singleton limiter，这个组合会静默丢失 tick。
	ErrOverlapWaitWithConcurrency = errors.New("job: overlap wait cannot be combined with a global concurrency limit")

	// ErrInvalidCron 表示 cron 表达式无法解析或没有未来触发时间。
	ErrInvalidCron = errors.New("job: invalid cron expression")

	// ErrEmptyHandlerName 表示注册了空名字的 handler。
	ErrEmptyHandlerName = errors.New("job: handler name must not be empty")

	// ErrNilHandler 表示注册了空 handler。
	ErrNilHandler = errors.New("job: handler must not be nil")

	// ErrDuplicateHandler 表示同一个名字注册了多次。
	ErrDuplicateHandler = errors.New("job: duplicate handler")

	// ErrHandlerNotFound 表示配置里声明的 job 没有对应的 handler。
	ErrHandlerNotFound = errors.New("job: configured job has no handler")

	// ErrJobNotConfigured 表示注册的 handler 在配置里找不到对应的 job。
	ErrJobNotConfigured = errors.New("job: registered handler is not configured")

	// ErrPanic 表示 handler 内部 panic 已被恢复。
	ErrPanic = errors.New("job: panic recovered")
)
