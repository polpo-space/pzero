package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestRetiredGeneratorConfiguration(t *testing.T) {
	orig, cfgFile, envFile := C, CfgFile, CfgEnvFile
	t.Cleanup(func() {
		C, CfgFile, CfgEnvFile = orig, cfgFile, envFile
		viper.Reset()
	})
	t.Chdir(t.TempDir())
	CfgFile, CfgEnvFile = ".pzero.yaml", ".pzero.env.yaml"
	for _, key := range []string{"gen.git-change", "gen.route2code", "gen.swagger.route2code"} {
		for _, source := range []string{"yaml", "env"} {
			t.Run(key+"/"+source, func(t *testing.T) {
				viper.Reset()
				C = Config{}
				if source == "yaml" {
					parts := strings.Split(key, ".")
					var body strings.Builder
					for i, part := range parts {
						body.WriteString(strings.Repeat("  ", i) + part + ":")
						if i == len(parts)-1 {
							body.WriteString(" false")
						}
						body.WriteByte('\n')
					}
					if err := os.WriteFile(CfgFile, []byte(body.String()), 0o600); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = os.Remove(CfgFile) })
				} else {
					t.Setenv("PZERO_"+strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(key)), "false")
				}
				err := InitConfig(&cobra.Command{Use: "pzero"})
				if err == nil || !strings.Contains(err.Error(), key+" has been retired") {
					t.Fatalf("expected migration error for %s, got %v", key, err)
				}
			})
		}
	}
}

func TestTraverseCommandsBindsLocalFlagsOnly(t *testing.T) {
	orig := C
	t.Cleanup(func() {
		C = orig
		viper.Reset()
	})
	C = Config{}
	viper.Reset()

	root := &cobra.Command{Use: "pzero"}
	root.PersistentFlags().String("style", "go_zero", "")

	gen := &cobra.Command{Use: "gen"}
	gen.PersistentFlags().StringSlice("model-datasource-url", []string{}, "")
	gen.Flags().StringSlice("desc", []string{}, "")
	root.AddCommand(gen)

	swagger := &cobra.Command{Use: "swagger"}
	swagger.Flags().String("output", "desc/swagger", "")
	gen.AddCommand(swagger)

	model := &cobra.Command{
		Use: "model",
		RunE: func(cmd *cobra.Command, args []string) error {
			return viper.Unmarshal(&C)
		},
	}
	gen.AddCommand(model)

	if err := TraverseCommands("", root); err != nil {
		t.Fatal(err)
	}

	if viper.IsSet("gen.swagger.style") {
		t.Fatal("inherited root flag must not bind as gen.swagger.style")
	}
	if viper.IsSet("gen.model.model-datasource-url") {
		t.Fatal("inherited model flag must not bind as gen.model.model-datasource-url")
	}

	root.SetArgs([]string{"gen", "model", "--model-datasource-url", "postgres://x"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(C.Gen.ModelDatasourceUrl) != 1 || C.Gen.ModelDatasourceUrl[0] != "postgres://x" {
		t.Fatalf("model-datasource-url = %v", C.Gen.ModelDatasourceUrl)
	}
}

func TestTraverseCommandsSharedModelFlagsStayOnGenKeys(t *testing.T) {
	orig := C
	t.Cleanup(func() {
		C = orig
		viper.Reset()
	})
	C = Config{}
	viper.Reset()

	root := &cobra.Command{Use: "pzero"}

	modelFlags := pflag.NewFlagSet("model", pflag.ContinueOnError)
	modelFlags.StringSlice("model-datasource-url", []string{}, "")

	gen := &cobra.Command{Use: "gen"}
	gen.Flags().AddFlagSet(modelFlags)
	root.AddCommand(gen)

	swagger := &cobra.Command{Use: "swagger"}
	gen.AddCommand(swagger)

	model := &cobra.Command{
		Use: "model",
		RunE: func(cmd *cobra.Command, args []string) error {
			return viper.Unmarshal(&C)
		},
	}
	model.Flags().AddFlagSet(modelFlags)
	gen.AddCommand(model)

	if err := TraverseCommands("", root); err != nil {
		t.Fatal(err)
	}

	root.SetArgs([]string{"gen", "model", "--model-datasource-url", "postgres://shared"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(C.Gen.ModelDatasourceUrl) != 1 || C.Gen.ModelDatasourceUrl[0] != "postgres://shared" {
		t.Fatalf("shared model flags must unmarshal to gen.model-datasource-url, got %v", C.Gen.ModelDatasourceUrl)
	}
	if swagger.Flags().Lookup("model-datasource-url") != nil {
		t.Fatal("swagger must not inherit model flags")
	}
}
