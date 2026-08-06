{{ if has "job" .Features }}package server

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"

	"{{ .Module }}/internal/job"
	"{{ .Module }}/internal/svc"
)

// JobServer runs scheduled jobs in-process with the RPC server.
type JobServer struct {
	svcCtx *svc.ServiceContext
	cron   *cron.Cron
}

func NewJobServer(svcCtx *svc.ServiceContext) *JobServer {
	c := cron.New(cron.WithSeconds())
	exampleJob := job.NewExampleJob(svcCtx)

	_, _ = c.AddFunc("@every 5s", func() {
		if err := exampleJob.EveryFiveSeconds(context.Background()); err != nil {
			logx.Errorf("job EveryFiveSeconds failed: %v", err)
		}
	})

	_, _ = c.AddFunc("0 * * * * *", func() {
		if err := exampleJob.EveryMinute(context.Background()); err != nil {
			logx.Errorf("job EveryMinute failed: %v", err)
		}
	})

	return &JobServer{
		svcCtx: svcCtx,
		cron:   c,
	}
}

// Start implements service.Service. cron.Start is async; Stop waits for jobs to finish.
func (s *JobServer) Start() {
	logx.Info("Starting job server...")
	s.cron.Start()
}

// Stop implements service.Service.
func (s *JobServer) Stop() {
	logx.Info("Stopping job server...")
	ctx := s.cron.Stop()
	<-ctx.Done()
}
{{ end -}}
