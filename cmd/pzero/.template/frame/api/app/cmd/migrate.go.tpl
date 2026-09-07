{{ if has "model" .Features }}package cmd

import (
	migratorcmd "github.com/polpo-space/pzero/runtime/migrator/cmd"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"{{ .Module }}/internal/config"
)

func init() {
	rootCmd.AddCommand(migratorcmd.NewCommand(func(configPath string) (sqlx.SqlConf, error) {
		var c config.Config
		if err := conf.Load(configPath, &c, conf.UseEnv()); err != nil {
			return sqlx.SqlConf{}, err
		}
		return c.Sqlx.SqlConf, nil
	}))
}
{{ end -}}
