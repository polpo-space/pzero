{{ if has "job" .Features }}package server

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"

	"{{ .Module }}/internal/config"
	"{{ .Module }}/internal/job"
	"{{ .Module }}/internal/svc"
)

// JobServer runs scheduled jobs in-process with the RPC server.
// Schedules come from config (job.jobs); handlers are wired by name here.
type JobServer struct {
	svcCtx  *svc.ServiceContext
	cron    *cron.Cron
	workers chan struct{}
}

func NewJobServer(svcCtx *svc.ServiceContext) *JobServer {
	cfg := svcCtx.Config.Job
	workers := cfg.Workers
	if workers <= 0 {
		workers = 1
	}

	opts := []cron.Option{cron.WithSeconds()}
	if cfg.Timezone != "" {
		loc, err := time.LoadLocation(cfg.Timezone)
		logx.Must(err)
		opts = append(opts, cron.WithLocation(loc))
	}

	s := &JobServer{
		svcCtx:  svcCtx,
		cron:    cron.New(opts...),
		workers: make(chan struct{}, workers),
	}

	exampleJob := job.NewExampleJob(svcCtx)
	handlers := map[string]func(context.Context) error{
		"exampleInterval": exampleJob.ExampleInterval,
		"exampleMinute":   exampleJob.ExampleMinute,
	}
	s.registerJobs(cfg, handlers)
	return s
}

func (s *JobServer) registerJobs(cfg config.JobConf, handlers map[string]func(context.Context) error) {
	for name := range cfg.Jobs {
		if _, ok := handlers[name]; !ok {
			logx.Must(fmt.Errorf("job %s has no handler", name))
		}
	}

	for name, spec := range cfg.Jobs {
		if !spec.Enable {
			continue
		}
		if spec.Cron == "" {
			logx.Must(fmt.Errorf("job %s has empty cron", name))
		}

		jobName, jobHandler := name, handlers[name]
		_, err := s.cron.AddFunc(spec.Cron, func() {
			s.run(jobName, jobHandler)
		})
		logx.Must(err)
		logx.Infof("job registered: %s cron=%s", jobName, spec.Cron)
	}
}

func (s *JobServer) run(name string, handler func(context.Context) error) {
	s.workers <- struct{}{}
	defer func() { <-s.workers }()

	if err := handler(context.Background()); err != nil {
		logx.Errorf("job %s failed: %v", name, err)
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
