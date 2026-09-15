package gen

import (
	"strings"
	"testing"
)

func TestRetiredGeneratorFlags(t *testing.T) {
	gen := GetCommand()
	for _, cmd := range append(gen.Commands(), gen) {
		for _, name := range []string{"git-change", "route2code"} {
			if err := cmd.ParseFlags([]string{"--" + name}); err == nil || !strings.Contains(err.Error(), "unknown flag") {
				t.Fatalf("%s --%s must be rejected, got %v", cmd.CommandPath(), name, err)
			}
		}
	}
}

func TestGetCommandIncludesModel(t *testing.T) {
	cmd := GetCommand()
	for _, c := range cmd.Commands() {
		if c.Name() == "model" {
			return
		}
	}
	t.Fatal("expected gen model subcommand")
}
