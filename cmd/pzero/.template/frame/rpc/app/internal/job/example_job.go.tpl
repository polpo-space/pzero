{{ if has "job" .Features }}package job

import (
	"context"

	joblogic "{{ .Module }}/internal/logic/job"
	"{{ .Module }}/internal/svc"
)

// ExampleJob is a thin job handler that delegates to logic.
// Handler names must match keys under job.jobs in etc.yaml.
type ExampleJob struct {
	svcCtx *svc.ServiceContext
}

func NewExampleJob(svcCtx *svc.ServiceContext) *ExampleJob {
	return &ExampleJob{svcCtx: svcCtx}
}

func (j *ExampleJob) ExampleInterval(ctx context.Context) error {
	return joblogic.NewExampleLogic(ctx, j.svcCtx).ExampleInterval()
}

func (j *ExampleJob) ExampleMinute(ctx context.Context) error {
	return joblogic.NewExampleLogic(ctx, j.svcCtx).ExampleMinute()
}
{{ end -}}
