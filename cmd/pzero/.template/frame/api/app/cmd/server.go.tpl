package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"

	"{{ .Module }}/internal/config"
	"{{ .Module }}/internal/middleware"
	"{{ .Module }}/internal/handler"
	"{{ .Module }}/internal/svc"
)

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

		// create rest server
		restServer := rest.MustNewServer(c.Rest.RestConf)

		// register auto generated routes
		handler.RegisterHandlers(restServer, svcCtx)
		// register middleware
		middleware.Register(restServer)

		group := service.NewServiceGroup()
		group.Add(restServer)

        logx.Infof("Starting rest server at %s:%d...", c.Rest.Host, c.Rest.Port)
		group.Start()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
