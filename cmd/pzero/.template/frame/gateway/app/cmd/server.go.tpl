package cmd

import (
    "github.com/common-nighthawk/go-figure"
	{{ if has "model" .Features }}
    "github.com/polpo-space/pzero/core/stores/migrate"{{end}}
	"github.com/polpo-space/pzero/core/swaggerv2"
	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/gateway"
	"github.com/zeromicro/go-zero/zrpc"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"


	"{{ .Module }}/desc/pb"
	"{{ .Module }}/internal/config"
	"{{ .Module }}/internal/custom"
	"{{ .Module }}/internal/global"
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

    	// print banner
        printBanner(c.Banner)
        // print version
        printVersion()

		{{ if has "model" .Features }}m, err := migrate.NewMigrate(c.Sqlx.SqlConf)
        logx.Must(err)
        defer m.Close()
        logx.Must(m.Up()){{end}}

        // create service context
		svcCtx := svc.NewServiceContext(c)

    	// write protoSets to local
        protoSets, err := pb.WriteToLocal(pb.Embed)
        logx.Must(err)
        c.Gateway.Upstreams[0].ProtoSets = protoSets
		svcCtx.Config = c

        // create zrpc server
        zrpcServer := zrpc.MustNewServer(c.Zrpc.RpcServerConf, func(grpcServer *grpc.Server) {
        	server.RegisterZrpcServer(grpcServer, svcCtx)
               {{if not .Serverless }}// register plugins
               plugins.LoadPlugins(grpcServer, svcCtx){{end}}
			if c.Zrpc.Mode == service.DevMode || c.Zrpc.Mode == service.TestMode {
        		reflection.Register(grpcServer)
        	}
        })
        // create gateway server
        gatewayServer := gateway.MustNewServer(svcCtx.Config.Gateway.GatewayConf, middleware.WithHeaderProcessor())
        // register swagger routes
        swaggerv2.RegisterRoutes(gatewayServer.Server)
        // // create custom server
        customServer := custom.New()

        // register middleware
        middleware.Register(zrpcServer, gatewayServer)

        group := service.NewServiceGroup()
        group.Add(zrpcServer)
        group.Add(gatewayServer)
        group.Add(customServer)

        logx.Infof("Starting rpc server at %s...", c.Zrpc.ListenOn)
        logx.Infof("Starting gateway server at %s:%d...", svcCtx.Config.Gateway.Host, svcCtx.Config.Gateway.Port)
        group.Start()
	},
}

func printBanner(c config.BannerConf) {
	figure.NewColorFigure(c.Text, c.FontName, c.Color, true).Print()
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
