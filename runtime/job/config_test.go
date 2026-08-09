package job

import (
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigNormalizeDefaults(t *testing.T) {
	cfg, location, err := Config{
		Jobs: map[string]Spec{
			"foo": {Enable: true, Cron: "0 * * * * *"},
		},
	}.normalize()
	require.NoError(t, err)

	assert.Equal(t, time.Local, location)
	assert.Equal(t, defaultShutdownTimeout, cfg.ShutdownTimeout)
	assert.Equal(t, LimitModeWait, cfg.Concurrency.Mode)
	assert.Zero(t, cfg.Concurrency.Limit, "全局并发默认不限制")
	assert.Equal(t, OverlapSkip, cfg.Jobs["foo"].Overlap)
}

func TestConfigNormalizeTimezone(t *testing.T) {
	cfg, location, err := Config{
		Timezone: "Asia/Shanghai",
		Jobs:     map[string]Spec{"foo": {Cron: "@every 5s"}},
	}.normalize()
	require.NoError(t, err)
	assert.Equal(t, "Asia/Shanghai", location.String())
	assert.Equal(t, "Asia/Shanghai", cfg.Timezone)

	_, _, err = Config{Timezone: "Mars/Olympus"}.normalize()
	assert.ErrorContains(t, err, "load timezone")
}

func TestConfigNormalizeInvalid(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name:    "negative concurrency limit",
			cfg:     Config{Concurrency: ConcurrencyConf{Limit: -1}},
			wantErr: ErrInvalidConcurrencyLimit,
		},
		{
			name:    "unknown concurrency mode",
			cfg:     Config{Concurrency: ConcurrencyConf{Mode: "queue"}},
			wantErr: ErrInvalidLimitMode,
		},
		{
			name:    "cron and every together",
			cfg:     Config{Jobs: map[string]Spec{"foo": {Cron: "* * * * *", Every: time.Second}}},
			wantErr: ErrScheduleConflict,
		},
		{
			name:    "no schedule",
			cfg:     Config{Jobs: map[string]Spec{"foo": {}}},
			wantErr: ErrScheduleRequired,
		},
		{
			name:    "negative every",
			cfg:     Config{Jobs: map[string]Spec{"foo": {Every: -time.Second}}},
			wantErr: ErrInvalidEvery,
		},
		{
			name:    "negative timeout",
			cfg:     Config{Jobs: map[string]Spec{"foo": {Cron: "* * * * *", Timeout: -time.Second}}},
			wantErr: ErrInvalidTimeout,
		},
		{
			name:    "unknown overlap",
			cfg:     Config{Jobs: map[string]Spec{"foo": {Cron: "* * * * *", Overlap: "queue"}}},
			wantErr: ErrInvalidOverlap,
		},
		{
			name: "overlap wait with global concurrency",
			cfg: Config{
				Concurrency: ConcurrencyConf{Limit: 2},
				Jobs:        map[string]Spec{"foo": {Cron: "* * * * *", Overlap: OverlapWait}},
			},
			wantErr: ErrOverlapWaitWithConcurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := tt.cfg.normalize()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// 禁用的 job 同样要校验，否则把 enable 翻成 true 才发现配置有问题。
func TestConfigNormalizeValidatesDisabledJobs(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr error
	}{
		{
			name:    "missing schedule",
			spec:    Spec{Enable: false},
			wantErr: ErrScheduleRequired,
		},
		{
			name:    "invalid cron",
			spec:    Spec{Enable: false, Cron: "not a cron"},
			wantErr: ErrInvalidCron,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := Config{Jobs: map[string]Spec{"foo": tt.spec}}.normalize()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestLimitModeTranslation(t *testing.T) {
	mode, err := LimitModeSkip.gocronMode()
	require.NoError(t, err)
	assert.Equal(t, gocron.LimitModeReschedule, mode)

	mode, err = LimitModeWait.gocronMode()
	require.NoError(t, err)
	assert.Equal(t, gocron.LimitModeWait, mode)
}

func TestOverlapTranslation(t *testing.T) {
	_, singleton, err := OverlapAllow.singletonMode()
	require.NoError(t, err)
	assert.False(t, singleton)

	mode, singleton, err := OverlapSkip.singletonMode()
	require.NoError(t, err)
	assert.True(t, singleton)
	assert.Equal(t, gocron.LimitModeReschedule, mode)

	mode, singleton, err = OverlapWait.singletonMode()
	require.NoError(t, err)
	assert.True(t, singleton)
	assert.Equal(t, gocron.LimitModeWait, mode)
}

// 现有脚手架使用 6 位含秒的 cron 与 @every 描述符，两者都必须继续可用。
func TestSpecDefinitionAcceptsSecondsAndDescriptor(t *testing.T) {
	for _, crontab := range []string{"0 */5 * * * *", "0 * * * *", "@every 5s", "@daily"} {
		t.Run(crontab, func(t *testing.T) {
			scheduler, err := gocron.NewScheduler()
			require.NoError(t, err)
			t.Cleanup(func() { _ = scheduler.Shutdown() })

			_, err = scheduler.NewJob(
				Spec{Cron: crontab}.definition(),
				gocron.NewTask(func() {}),
			)
			assert.NoError(t, err)
		})
	}
}

func TestValidateCronMatchesGocronSyntax(t *testing.T) {
	tests := []struct {
		name    string
		cron    string
		wantErr bool
	}{
		{name: "five fields", cron: "*/5 * * * *"},
		{name: "six fields", cron: "0 */5 * * * *"},
		{name: "descriptor", cron: "@every 5s"},
		{name: "explicit timezone", cron: "CRON_TZ=UTC 0 * * * * *"},
		{name: "invalid", cron: "not a cron", wantErr: true},
		{name: "no future run", cron: "0 0 0 31 2 *", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCron(tt.cron, time.UTC, time.Now())

			scheduler, schedulerErr := gocron.NewScheduler(gocron.WithLocation(time.UTC))
			require.NoError(t, schedulerErr)
			t.Cleanup(func() { _ = scheduler.Shutdown() })
			_, gocronErr := scheduler.NewJob(
				gocron.CronJob(tt.cron, true),
				gocron.NewTask(func() {}),
			)
			assert.Equal(t, gocronErr != nil, err != nil, "validator must match gocron acceptance")

			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidCron)
				return
			}
			assert.NoError(t, err)
		})
	}
}
