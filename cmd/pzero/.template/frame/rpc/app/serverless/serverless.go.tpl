{{ if .Serverless }}package serverless

import (
	"path/filepath"

	"github.com/zeromicro/go-zero/core/conf"
	"google.golang.org/grpc"

	"{{ .Module }}/internal/config"
	"{{ .Module }}/internal/server"
	"{{ .Module }}/internal/svc"
)

type Serverless struct {
	SvcCtx             *svc.ServiceContext // 服务上下文
	RegisterZrpcServer func(grpcServer *grpc.Server, ctx *svc.ServiceContext)
}

func New() *Serverless {
	var c config.Config
	conf.MustLoad(filepath.Join("plugins", "{{ .DirName }}", "etc", "etc.yaml"), &c, conf.UseEnv())

	svcCtx := svc.NewServiceContext(c)

	return &Serverless{
		SvcCtx:             svcCtx,
		RegisterZrpcServer: server.RegisterZrpcServer,
	}
}{{end}}
