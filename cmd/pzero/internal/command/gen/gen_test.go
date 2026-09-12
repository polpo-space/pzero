package gen

import "testing"

func TestGetCommandIncludesModel(t *testing.T) {
	cmd := GetCommand()
	for _, c := range cmd.Commands() {
		if c.Name() == "model" {
			return
		}
	}
	t.Fatal("expected gen model subcommand")
}
