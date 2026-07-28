/*
Copyright © 2024 jaronnie <jaron@jaronnie.com>

*/

package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	goversion "github.com/hashicorp/go-version"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"

	"github.com/polpo-space/pzero/cmd/pzero/internal/command/add"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/check"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/completion"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/format"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/gen"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/migrate"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/new"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/serverless"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/template"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/upgrade"
	versioncmd "github.com/polpo-space/pzero/cmd/pzero/internal/command/version"
	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/desc"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
	"github.com/polpo-space/pzero/cmd/pzero/internal/hooks"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console"
)

var WorkingDir string

// embeded
var (
	//go:embed all:.template
	Template embed.FS
)

var (
	version = "v1.4.0"
	commit  string
	date    string
)

func main() {
	embeded.Template = Template
	versioncmd.Version = version
	versioncmd.Date = date
	versioncmd.Commit = commit

	Execute()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "pzero",
	Short: `Used to create project by templates and generate server/client code by api/proto/sql file.
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Display logo
		if os.Getenv("PZERO_HOOK_TRIGGERED") != "true" && os.Getenv("PZERO_FORKED") != "true" && !config.C.Quiet {
			console.DisplayLogo(version, lo.If(config.C.Debug, func() []string {
				var toolVersion []string
				tv := config.C.ToolVersion()

				toolVersion = appendToolVersion(toolVersion, "goctl", tv.GoctlVersion)

				frameType, err := desc.GetFrameType()
				cobra.CheckErr(err)

				if frameType == "rpc" || frameType == "gateway" {
					toolVersion = appendToolVersion(toolVersion, "protoc", tv.ProtocVersion)
					toolVersion = appendToolVersion(toolVersion, "protoc-gen-go", tv.ProtocGenGoVersion)
					toolVersion = appendToolVersion(toolVersion, "protoc-gen-go-grpc", tv.ProtocGenGoGrpcVersion)
					toolVersion = appendToolVersion(toolVersion, "protoc-gen-openapiv2", tv.ProtocGenOpenapiv2Version)
				}
				return toolVersion
			}()).Else(nil))
		}

		// Run environment check first
		if cmd.Name() != check.GetCommand().Use && cmd.Name() != versioncmd.GetCommand().Use {
			frameType, err := desc.GetFrameType()
			if err != nil {
				return err
			}
			if frameType != "" {
				if err := check.RunCheck(false); err != nil {
					return err
				}
			}
		}

		// Check home
		if !pathx.FileExists(config.C.Home) {
			if pathx.FileExists(filepath.Join(config.C.HomeDir(), ".pzero", "templates", versioncmd.Version)) {
				config.C.Home = filepath.Join(config.C.HomeDir(), ".pzero", "templates", versioncmd.Version)
				embeded.Home = config.C.Home
			} else {
				config.C.Home = ""
			}
		} else {
			embeded.Home = config.C.Home
		}

		if os.Getenv("PZERO_HOOK_TRIGGERED") != "true" && os.Getenv("PZERO_FORKED") != "true" && config.C.Debug && !config.C.Quiet {
			cfgJSON, err := config.MarshalDebugJSONForCommand(config.C, cmd)
			if err != nil {
				return err
			}
			title := console.Green("Parse") + " " + console.Yellow("Config")
			if err := console.DisplayJSONBox(title, cfgJSON); err != nil {
				return err
			}
		}

		return hooks.Run(cmd, "Before", "global", config.C.Hooks.Before)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if parseBool, err := strconv.ParseBool(cmd.Flags().Lookup("version").Value.String()); err == nil && parseBool {
			versioncmd.GetVersion()
			return
		}
		if err := cmd.Help(); err != nil {
			cobra.CheckErr(err)
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return hooks.Run(cmd, "After", "global", config.C.Hooks.After)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if console.IsRenderedError(err) {
			os.Exit(1)
		}
		cobra.CheckErr(err)
	}
}

func appendToolVersion(items []string, name string, v *goversion.Version) []string {
	if v == nil {
		return items
	}

	return append(items, fmt.Sprintf("%s v%s", name, v.String()))
}

func init() {
	cobra.OnInitialize(InitConfig)

	rootCmd.PersistentFlags().StringP("style", "", "gozero", "The file naming format, see [https://github.com/zeromicro/go-zero/blob/master/tools/goctl/config/readme.md]")
	rootCmd.PersistentFlags().StringP("home", "", ".template", "set template home")
	rootCmd.PersistentFlags().StringVarP(&config.CfgFile, "config", "f", ".pzero.yaml", "set config file")
	rootCmd.PersistentFlags().StringVarP(&config.CfgEnvFile, "config-env", "", ".pzero.env.yaml", "set config env file")
	rootCmd.PersistentFlags().BoolP("debug", "", false, "debug mode")
	rootCmd.PersistentFlags().BoolP("quiet", "", false, "quiet mode")
	rootCmd.PersistentFlags().IntP("debug-sleep-time", "", 0, "debug sleep time")
	rootCmd.Flags().BoolP("version", "v", false, "show version")
	rootCmd.PersistentFlags().StringVarP(&WorkingDir, "working-dir", "w", "", "set working directory")
	rootCmd.PersistentFlags().StringSliceP("register-tpl-val", "", []string{}, "register tpl value, e.g. --register-tpl-val key=value")

	rootCmd.AddCommand(check.GetCommand())
	rootCmd.AddCommand(completion.GetCommand())
	rootCmd.AddCommand(format.GetCommand())
	rootCmd.AddCommand(add.GetCommand())
	rootCmd.AddCommand(gen.GetCommand())
	rootCmd.AddCommand(migrate.GetCommand())
	rootCmd.AddCommand(new.GetCommand())
	rootCmd.AddCommand(serverless.GetCommand())
	rootCmd.AddCommand(template.GetCommand())
	rootCmd.AddCommand(upgrade.GetCommand())
	rootCmd.AddCommand(versioncmd.GetCommand())
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	if len(os.Args) >= 2 {
		if os.Args[1] == versioncmd.GetCommand().Use {
			return
		}
	}

	if WorkingDir != "" && os.Getenv("PZERO_FORKED") != "true" {
		// Convert relative path to absolute path
		absPath, err := filepath.Abs(WorkingDir)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("failed to get absolute path for working directory: %w", err))
		}

		// Verify the directory exists
		if _, err := os.Stat(absPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				cobra.CheckErr(fmt.Errorf("working directory does not exist: %s", absPath))
			}
			cobra.CheckErr(fmt.Errorf("failed to access working directory: %w", err))
		}

		if err := os.Chdir(absPath); err != nil {
			cobra.CheckErr(fmt.Errorf("failed to change to working directory: %w", err))
		}
	}

	cobra.CheckErr(config.InitConfig(rootCmd))

	logx.Disable()
	if config.C.Debug {
		time.Sleep(time.Duration(config.C.DebugSleepTime) * time.Second)
	}
}
