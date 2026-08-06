package cmd

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"{{ .Module }}/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "{{ .APP }} version",
	Long:  `{{ .APP }} version`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func printVersion() {
	var versionBuffer bytes.Buffer

	versionBuffer.WriteString(fmt.Sprintf("{{ .APP }} version %s %s/%s\n", version.Version, runtime.GOOS, runtime.GOARCH))
	versionBuffer.WriteString(fmt.Sprintf("Go version %s\n", runtime.Version()))
	versionBuffer.WriteString(fmt.Sprintf("Git commit %s\n", version.Commit))
	versionBuffer.WriteString(fmt.Sprintf("Build date: %s\n", version.Date))

	fmt.Print(versionBuffer.String())
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
