package gen

import (
	"testing"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
)

func TestShouldRunModelStage(t *testing.T) {
	orig := config.C
	t.Cleanup(func() { config.C = orig })

	t.Run("empty desc", func(t *testing.T) {
		config.C = config.Config{}
		if !shouldRunModelStage() {
			t.Fatal("full gen without desc must run model")
		}
	})

	t.Run("api desc skips model", func(t *testing.T) {
		config.C = config.Config{}
		config.C.Gen.Desc = []string{"desc/api/user.api"}
		if shouldRunModelStage() {
			t.Fatal("api desc must skip model in full gen")
		}
	})

	t.Run("sql desc still enters model", func(t *testing.T) {
		config.C = config.Config{}
		config.C.Gen.Desc = []string{"desc/sql/users.sql"}
		if !shouldRunModelStage() {
			t.Fatal("sql desc must reach genmodel so it can reject snapshot input")
		}
	})
}
