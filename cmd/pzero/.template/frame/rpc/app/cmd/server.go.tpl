package cmd

import (
	"github.com/spf13/cobra"
    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
    "github.com/zeromicro/go-zero/zrpc"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

	"{{ .Module }}/internal/config"
	"{{ .Module }}/internal/middleware"
	"{{ .Module }}/internal/server"
	"{{ .Module }}/internal/svc"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "{{ .APP }} server",
	Long:  "{{ .APP }} server",
	Run: func(cmd *cobra.Command, args []string) {
		var c config.Config
		conf.MustLoad(cmd.Flag("config").Value.String(), &c, conf.UseEnv())

        // set up logger
        logx.Must(logx.SetUp(c.Log.LogConf))

	    printVersion()

		// create service context
		svcCtx := svc.NewServiceContext(c)
        // create zrpc server
	    zrpcServer := zrpc.MustNewServer(c.Zrpc.RpcServerConf, func(grpcServer *grpc.Server) {
            server.RegisterZrpcServer(grpcServer, svcCtx)
            if c.Zrpc.Mode == service.DevMode || c.Zrpc.Mode == service.TestMode {
            	reflection.Register(grpcServer)
            }
        })
        // register middleware
        middleware.Register(zrpcServer)

	    group := service.NewServiceGroup()
	    group.Add(zrpcServer)
{{ if has "job" .Features }}
	    if c.Job.Enable {
	    	group.Add(server.NewJobServer(svcCtx))
	    	logx.Info("Job server enabled")
	    }
{{ end }}
        logx.Infof("Starting rpc server at %s...", c.Zrpc.ListenOn)
        group.Start()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
