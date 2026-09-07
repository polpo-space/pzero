{{ if has "job" .Features }}package server

import (
	runtimejob "github.com/polpo-space/pzero/runtime/job"
	"github.com/zeromicro/go-zero/core/logx"

	"{{ .Module }}/internal/job"
	"{{ .Module }}/internal/svc"
)

// NewJobServer 构建定时任务调度器。调度、并发、超时、panic 恢复与优雅退出
// 都在 pzero/runtime/job 里，这里只负责把配置和 handler 接上。
//
// 配置里的 job 与注册的 handler 必须严格一一对应，任何一边缺失都会导致
// 启动失败。
func NewJobServer(svcCtx *svc.ServiceContext) *runtimejob.Server {
	server, err := runtimejob.New(
		svcCtx.Config.Job,
		job.Registry(svcCtx),
		runtimejob.WithNamespace(svcCtx.Config.Zrpc.Name),
	)
	logx.Must(err)

	return server
}
{{ end -}}
