{{ if .Serverless }}package serverless

import (
	"path/filepath"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"

	"{{ .Module }}/internal/config"
	"{{ .Module }}/desc/pb"
	"{{ .Module }}/internal/server"
	"{{ .Module }}/internal/svc"
)

type Serverless struct {
	SvcCtx             *svc.ServiceContext // 服务上下文
	RegisterZrpcServer func(grpcServer *grpc.Server, ctx *svc.ServiceContext)
	ProtoSets          []string
}

func New() *Serverless {
	var c config.Config
	conf.MustLoad(filepath.Join("plugins", "{{ .DirName }}", "etc", "etc.yaml"), &c, conf.UseEnv())

	svcCtx := svc.NewServiceContext(c)

	// get protoSets
	protoSets, err := pb.WriteToLocal(pb.Embed)
	logx.Must(err)

	return &Serverless{
		SvcCtx:             svcCtx,
		RegisterZrpcServer: server.RegisterZrpcServer,
		ProtoSets:          protoSets,
	}
}{{end}}
