/*
Copyright © 2024 jaronnie <jaron@jaronnie.com>

*/

package gen

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/polpo-space/pzero/cmd/pzero/internal/command/gen/gen"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/gen/genswagger"
	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/hooks"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: `Used to generate server code with api, proto, sql desc file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := gen.Run()
		if err == nil {
			return nil
		}

		_ = hooks.Run(cmd, "After", "gen", config.C.Gen.Hooks.After)
		if config.C.Quiet {
			return err
		}
		return console.MarkRenderedError(err)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// genSwaggerCmd represents the genSwagger command
var genSwaggerCmd = &cobra.Command{
	Use:   "swagger",
	Short: `Gen swagger json docs by proto and api file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return genswagger.Gen()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func GetCommand() *cobra.Command {
	{
		genCmd.Flags().StringSliceP("desc", "", []string{}, "set desc path")
		genCmd.Flags().StringSliceP("desc-ignore", "", []string{}, "set desc ignore path")
		genCmd.Flags().BoolP("git-change", "", false, "set is git change, if changes then generate code")
		genCmd.Flags().StringP("api-types-dir", "", filepath.Join("internal", "types"), "set generated api types dir, relative to the project root")
		genCmd.Flags().BoolP("route2code", "", false, "is generate route2code")
		genCmd.Flags().StringSliceP("proto-dir", "", []string{}, "RPC proto scan roots, default desc/proto")
		genCmd.Flags().StringSliceP("proto-include", "", []string{}, "proto include path")
		genCmd.Flags().StringP("model-driver", "", "postgres", "goctl model driver, postgres only")
		genCmd.Flags().BoolP("model-strict", "", false, "goctl model strict mode, see [https://go-zero.dev/docs/tutorials/cli/model]")
		genCmd.Flags().StringSliceP("model-ignore-columns", "", []string{"create_at", "created_at", "create_time", "update_at", "updated_at", "update_time"}, "ignore columns of postgres model")
		genCmd.Flags().StringP("model-schema", "", "", "model schema")
		genCmd.Flags().BoolP("model-datasource", "", false, "goctl datasource")
		genCmd.Flags().StringSliceP("model-datasource-url", "", []string{}, "goctl model datasource url")
		genCmd.Flags().StringSliceP("model-datasource-table", "", []string{"*"}, "goctl model datasource table")
		genCmd.Flags().BoolP("model-cache", "", false, "goctl model cache")
		genCmd.Flags().StringSliceP("model-cache-table", "", []string{"*"}, "goctl model cache tables")
		genCmd.Flags().StringP("model-cache-prefix", "", "cache", "goctl model cache prefix")
		genCmd.Flags().BoolP("rpc-client", "", false, "generate rpc client code")
	}

	{
		genCmd.AddCommand(genSwaggerCmd)

		genSwaggerCmd.Flags().StringSliceP("desc", "", []string{}, "set desc path")
		genSwaggerCmd.Flags().StringSliceP("desc-ignore", "", []string{}, "set desc ignore path")
		genSwaggerCmd.Flags().StringP("output", "o", filepath.Join("desc", "swagger"), "set swagger output dir")
		genSwaggerCmd.Flags().BoolP("route2code", "", false, "is generate route2code")
		genSwaggerCmd.Flags().BoolP("merge", "", true, "is merge muti swagger to one file")
	}

	return genCmd
}
