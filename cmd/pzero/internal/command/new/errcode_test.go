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

// TestRunGeneratesErrcode 验证 api/rpc 脚手架均生成统一错误码包, rpc 校验中间件改用业务码。
func TestRunGeneratesErrcode(t *testing.T) {
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

			embeded.Home = templateHome
			config.C.Style = config.DefaultStyle
			config.C.New = config.NewConfig{
				Module: "example.com/" + frame + "-svc",
				Output: projectDir,
				Frame:  frame,
			}

			require.NoError(t, Run(frame+"-svc", filepath.Join("frame", frame, "app")))

			content, err := os.ReadFile(filepath.Join(projectDir, "internal", "errcode", "errcode.go"))
			require.NoError(t, err)
			for _, expected := range []string{
				"package errcode",
				"github.com/polpo-space/pzero/core/status",
				"status.RegisterWithMessage(code, message)",
				"InvalidParam = register(",
				"NotFound     = register(",
				"Internal     = register(",
			} {
				assert.Contains(t, string(content), expected)
			}

			if frame == "api" {
				response, err := os.ReadFile(filepath.Join(projectDir, "internal", "middleware", "response.go"))
				require.NoError(t, err)
				assert.Contains(t, string(response), "fromError.Message()")
				assert.NotContains(t, string(response), "fromError.Error()")
			}

			if frame != "rpc" {
				return
			}
			validator, err := os.ReadFile(filepath.Join(projectDir, "internal", "middleware", "validator.go"))
			require.NoError(t, err)
			assert.Contains(t, string(validator), config.C.New.Module+"/internal/errcode")
			assert.Contains(t, string(validator), "status.ErrorMessage(errcode.InvalidParam")
			assert.NotContains(t, string(validator), "google.golang.org/grpc/codes")
		})
	}
}
