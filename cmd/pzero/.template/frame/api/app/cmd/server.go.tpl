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
	"github.com/zeromicro/go-zero/rest"

	"{{ .Module }}/internal/config"
	"{{ .Module }}/internal/custom"
	"{{ .Module }}/internal/middleware"
	"{{ .Module }}/internal/handler"
	"{{ .Module }}/internal/svc"
	{{ if not .Serverless }}"{{ .Module }}/plugins"{{end}}
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

		// create rest server
		restServer := rest.MustNewServer(c.Rest.RestConf)
		// create custom server
		customServer := custom.New()

		// register auto generated routes
		handler.RegisterHandlers(restServer, svcCtx)
		// register swagger routes
		swaggerv2.RegisterRoutes(restServer)
		// register middleware
		middleware.Register(restServer)

	    {{ if not .Serverless }}// load plugins
	    plugins.LoadPlugins(restServer, svcCtx){{end}}

		group := service.NewServiceGroup()
		group.Add(restServer)
		group.Add(customServer)

        logx.Infof("Starting rest server at %s:%d...", c.Rest.Host, c.Rest.Port)
		group.Start()
	},
}

func printBanner(c config.BannerConf) {
	figure.NewColorFigure(c.Text, c.FontName, c.Color, true).Print()
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
