{{ if has "job" .Features }}package joblogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"{{ .Module }}/internal/svc"
)

// ExampleLogic holds job business logic.
type ExampleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExampleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExampleLogic {
	return &ExampleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ExampleLogic) EveryFiveSeconds() error {
	l.Info("example job: every five seconds")
	return nil
}

func (l *ExampleLogic) EveryMinute() error {
	l.Info("example job: every minute")
	return nil
}
{{ end -}}
