package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

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
