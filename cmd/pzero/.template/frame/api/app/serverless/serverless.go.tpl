{{ if .Serverless }}package serverless

import (
	"path/filepath"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"{{ .Module }}/internal/config"
    "{{ .Module }}/internal/handler"
    "{{ .Module }}/internal/svc"
)

type Serverless struct {
	SvcCtx        *svc.ServiceContext                                   // 服务上下文
	HandlerFunc   func(server *rest.Server, svcCtx *svc.ServiceContext) // 服务路由
}

// New serverless function
func New() *Serverless {
	var c config.Config
	conf.MustLoad(filepath.Join("plugins", "{{ .DirName }}", "etc", "etc.yaml"), &c, conf.UseEnv())

	svcCtx := svc.NewServiceContext(c)

	return &Serverless{
		SvcCtx:      svcCtx,
		HandlerFunc: handler.RegisterHandlers,
	}
}{{end}}
