package skillsinit

import (
	"fmt"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
)

func Run() error {
	err := embeded.WriteTemplateDir("skills", config.C.Skills.Init.Output)
	if err != nil {
		return fmt.Errorf("failed to initialized skills templates: %w", err)
	}

	if !config.C.Quiet {
		fmt.Printf("✓ Skills templates initialized successfully at: %s\n", config.C.Skills.Init.Output)
	}

	return nil
}
