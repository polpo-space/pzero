package cmd

import (
	"github.com/common-nighthawk/go-figure"
    {{ if has "model" .Features }}
	"github.com/polpo-space/pzero/core/stores/migrate"{{end}}
	"github.com/spf13/cobra"
    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
    "github.com/zeromicro/go-zero/zrpc"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

	"{{ .Module }}/internal/config"
	"{{ .Module }}/internal/custom"
	"{{ .Module }}/internal/middleware"
	"{{ .Module }}/internal/server"
	"{{ .Module }}/internal/svc"
	{{ if not .Serverless }}"{{ .Module }}/plugins"{{end}}
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

	    printBanner(c.Zrpc.Name)
	    printVersion()

		{{ if has "model" .Features }}m, err := migrate.NewMigrate(c.Sqlx.SqlConf)
        logx.Must(err)
        defer m.Close()
        logx.Must(m.Up()){{end}}

        // create service context
		svcCtx := svc.NewServiceContext(c)
        // create zrpc server
	    zrpcServer := zrpc.MustNewServer(c.Zrpc.RpcServerConf, func(grpcServer *grpc.Server) {
            server.RegisterZrpcServer(grpcServer, svcCtx)
                {{if not .Serverless }}// register plugins
                plugins.LoadPlugins(grpcServer, svcCtx){{end}}
            if c.Zrpc.Mode == service.DevMode || c.Zrpc.Mode == service.TestMode {
            	reflection.Register(grpcServer)
            }
        })
        // create custom server
	    customServer := custom.New()
        // register middleware
        middleware.Register(zrpcServer)

	    group := service.NewServiceGroup()
	    group.Add(zrpcServer)
	    group.Add(customServer)

        logx.Infof("Starting rpc server at %s...", c.Zrpc.ListenOn)
        group.Start()
	},
}

func printBanner(serviceName string) {
	figure.NewColorFigure(serviceName, "starwars", "green", false).Print()
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
