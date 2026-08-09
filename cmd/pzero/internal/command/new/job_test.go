package new

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
)

func TestRunGeneratesCompilableRPCJobScaffold(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	templateHome, err := filepath.Abs(filepath.Join(cwd, "../../..", ".template"))
	require.NoError(t, err)
	repoRoot, err := filepath.Abs(filepath.Join(cwd, "../../../../.."))
	require.NoError(t, err)

	oldConfig, oldHome := config.C, embeded.Home
	t.Cleanup(func() {
		config.C = oldConfig
		embeded.Home = oldHome
	})

	projectDir := filepath.Join(t.TempDir(), "billing-svc")
	embeded.Home = templateHome
	config.C.Style = config.DefaultStyle
	config.C.New = config.NewConfig{
		Module:   "example.com/billing-svc",
		Output:   projectDir,
		Frame:    "rpc",
		Features: []string{"job"},
	}

	require.NoError(t, Run("billing-svc", filepath.Join("frame", "rpc", "app")))

	for _, path := range []string{
		"internal/config/config.go",
		"internal/job/example_job.go",
		"internal/job/registry.go",
		"internal/logic/job/example_logic.go",
		"internal/server/job.go",
	} {
		assert.FileExists(t, filepath.Join(projectDir, path))
	}

	runGo(t, projectDir, "mod", "edit", "-replace", "github.com/polpo-space/pzero="+repoRoot)
	runGo(t, projectDir, "test", "-mod=mod",
		"./internal/config",
		"./internal/job",
		"./internal/logic/job",
		"./internal/server",
		"./internal/svc",
	)
}

func runGo(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", output)
}
