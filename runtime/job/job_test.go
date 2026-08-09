package job

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/service"
)

func noop(context.Context) error { return nil }

func TestBuildRegistryStrictMatch(t *testing.T) {
	cfg := Config{Jobs: map[string]Spec{"foo": {Cron: "* * * * *"}}}

	tests := []struct {
		name     string
		cfg      Config
		handlers []NamedHandler
		wantErr  error
	}{
		{
			name:     "matched",
			cfg:      cfg,
			handlers: []NamedHandler{Named("foo", noop)},
		},
		{
			name:     "configured job without handler",
			cfg:      Config{Jobs: map[string]Spec{"foo": {Cron: "* * * * *"}, "bar": {Cron: "* * * * *"}}},
			handlers: []NamedHandler{Named("foo", noop)},
			wantErr:  ErrHandlerNotFound,
		},
		{
			name:     "handler without configured job",
			cfg:      cfg,
			handlers: []NamedHandler{Named("foo", noop), Named("bar", noop)},
			wantErr:  ErrJobNotConfigured,
		},
		{
			name:     "duplicate handler",
			cfg:      cfg,
			handlers: []NamedHandler{Named("foo", noop), Named("foo", noop)},
			wantErr:  ErrDuplicateHandler,
		},
		{
			name:     "empty name",
			cfg:      cfg,
			handlers: []NamedHandler{Named("", noop)},
			wantErr:  ErrEmptyHandlerName,
		},
		{
			name:     "nil handler",
			cfg:      cfg,
			handlers: []NamedHandler{Named("foo", nil)},
			wantErr:  ErrNilHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildRegistry(tt.cfg, tt.handlers)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewRejectsInvalidCron(t *testing.T) {
	_, err := New(
		Config{Jobs: map[string]Spec{"foo": {Enable: true, Cron: "not a cron"}}},
		[]NamedHandler{Named("foo", noop)},
	)
	assert.ErrorContains(t, err, `job "foo"`)
}

func TestNewNamespacesJobNames(t *testing.T) {
	server, err := New(
		Config{Jobs: map[string]Spec{
			"beta":  {Enable: true, Cron: "0 0 * * * *"},
			"alpha": {Enable: true, Cron: "0 0 * * * *"},
			"gamma": {Enable: false, Cron: "0 0 * * * *"},
		}},
		[]NamedHandler{Named("alpha", noop), Named("beta", noop), Named("gamma", noop)},
		WithNamespace("payment.rpc"),
	)
	require.NoError(t, err)
	t.Cleanup(server.Stop)

	// 排序保证注册顺序确定，禁用的 job 不注册。
	assert.Equal(t, []string{"payment.rpc.alpha", "payment.rpc.beta"}, server.registered)
}

func TestServerImplementsService(t *testing.T) {
	server, err := New(
		Config{Jobs: map[string]Spec{"foo": {Enable: true, Cron: "0 0 * * * *"}}},
		[]NamedHandler{Named("foo", noop)},
	)
	require.NoError(t, err)

	var svc service.Service = server
	svc.Start()
	svc.Stop()
}

func TestServerRunsJob(t *testing.T) {
	var runs atomic.Int32
	server, err := New(
		Config{Jobs: map[string]Spec{"tick": {Enable: true, Every: 20 * time.Millisecond}}},
		[]NamedHandler{Named("tick", func(context.Context) error {
			runs.Add(1)
			return nil
		})},
	)
	require.NoError(t, err)

	server.Start()
	assert.Eventually(t, func() bool { return runs.Load() > 0 }, time.Second, 10*time.Millisecond)
	server.Stop()
}

// handler panic 必须被兜住并转成 error，不能带走进程。
func TestRunRecoversPanic(t *testing.T) {
	err := run(context.Background(), func(context.Context) error {
		panic("boom")
	}, 0)

	require.ErrorIs(t, err, ErrPanic)
	assert.ErrorContains(t, err, "boom")
	assert.ErrorContains(t, err, "job.TestRunRecoversPanic", "错误里应带上 panic 现场的调用栈")
}

func TestRunAppliesTimeout(t *testing.T) {
	err := run(context.Background(), func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}, 10*time.Millisecond)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRunWithoutTimeoutKeepsParentContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, func(ctx context.Context) error { return ctx.Err() }, 0)
	assert.ErrorIs(t, err, context.Canceled)
}

// Stop 会取消 job ctx，handler 才能感知进程正在退出。
func TestStopCancelsJobContext(t *testing.T) {
	canceled := make(chan struct{})
	started := make(chan struct{})

	server, err := New(
		Config{
			ShutdownTimeout: time.Second,
			Jobs: map[string]Spec{
				"blocking": {Enable: true, Every: 10 * time.Millisecond, Overlap: OverlapSkip},
			},
		},
		[]NamedHandler{Named("blocking", func(ctx context.Context) error {
			select {
			case <-started:
			default:
				close(started)
			}
			<-ctx.Done()
			close(canceled)
			return ctx.Err()
		})},
	)
	require.NoError(t, err)

	server.Start()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}

	server.Stop()
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("job context was not canceled on stop")
	}
}

func TestQualify(t *testing.T) {
	assert.Equal(t, "cleanup", qualify("", "cleanup"))
	assert.Equal(t, "order.rpc.cleanup", qualify("order.rpc", "cleanup"))
}
