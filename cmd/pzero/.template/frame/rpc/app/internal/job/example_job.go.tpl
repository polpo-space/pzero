{{ if has "job" .Features }}package job

import (
	"context"

	joblogic "{{ .Module }}/internal/logic/job"
	"{{ .Module }}/internal/svc"
)

// ExampleJob is a thin job handler that delegates to logic.
type ExampleJob struct {
	svcCtx *svc.ServiceContext
}

func NewExampleJob(svcCtx *svc.ServiceContext) *ExampleJob {
	return &ExampleJob{svcCtx: svcCtx}
}

// EveryFiveSeconds runs on a fixed interval.
func (j *ExampleJob) EveryFiveSeconds(ctx context.Context) error {
	return joblogic.NewExampleLogic(ctx, j.svcCtx).EveryFiveSeconds()
}

// EveryMinute runs on a cron schedule.
func (j *ExampleJob) EveryMinute(ctx context.Context) error {
	return joblogic.NewExampleLogic(ctx, j.svcCtx).EveryMinute()
}
{{ end -}}
