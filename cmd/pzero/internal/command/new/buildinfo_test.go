package new

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
)

func TestRunGeneratesInternalBuildInfo(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	templateHome, err := filepath.Abs(filepath.Join(cwd, "../../..", ".template"))
	require.NoError(t, err)

	oldConfig, oldHome := config.C, embeded.Home
	t.Cleanup(func() {
		config.C = oldConfig
		embeded.Home = oldHome
	})

	for _, frame := range []string{"api", "rpc"} {
		t.Run(frame, func(t *testing.T) {
			projectDir := filepath.Join(t.TempDir(), frame+"-svc")
			module := "example.com/" + frame + "-svc"

			embeded.Home = templateHome
			config.C.Style = config.DefaultStyle
			config.C.New = config.NewConfig{
				Module: module,
				Output: projectDir,
				Frame:  frame,
			}

			require.NoError(t, Run(frame+"-svc", filepath.Join("frame", frame, "app")))

			assert.FileExists(t, filepath.Join(projectDir, "internal", "buildinfo", "buildinfo.go"))
			assert.NoFileExists(t, filepath.Join(projectDir, "version", "version.go"))

			for _, path := range []string{
				filepath.Join("cmd", "version.go"),
				filepath.Join("internal", "logic", "version", "version.go"),
				"README.md",
			} {
				content, err := os.ReadFile(filepath.Join(projectDir, path))
				require.NoError(t, err)
				assert.Contains(t, string(content), module+"/internal/buildinfo")
				assert.NotContains(t, string(content), module+"/version")
			}
		})
	}
}
