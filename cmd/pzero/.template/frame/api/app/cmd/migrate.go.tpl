{{ if has "model" .Features }}package cmd

import (
	"github.com/polpo-space/pzero/runtime/pkg/migrator"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"{{ .Module }}/internal/config"
)

func init() {
	rootCmd.AddCommand(migrator.NewCommand(func(configPath string) (sqlx.SqlConf, error) {
		var c config.Config
		if err := conf.Load(configPath, &c, conf.UseEnv()); err != nil {
			return sqlx.SqlConf{}, err
		}
		return c.Sqlx.SqlConf, nil
	}))
}
{{ end -}}
