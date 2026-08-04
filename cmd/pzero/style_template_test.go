package main

import (
	"testing"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	pzerotemplatex "github.com/polpo-space/pzero/cmd/pzero/internal/pkg/templatex"
)

func TestAPIRPCConfigTemplatesWriteDefaultStyle(t *testing.T) {
	for _, frame := range []string{"api", "rpc"} {
		t.Run(frame, func(t *testing.T) {
			path := ".template/frame/" + frame + "/app/.pzero.yaml.tpl"
			content := readEmbeddedTemplate(t, path)
			rendered, err := pzerotemplatex.ParseTemplate(path, map[string]any{
				"Style": config.DefaultStyle,
			}, []byte(content))
			if err != nil {
				t.Fatalf("render %s config template: %v", frame, err)
			}
			if string(rendered) != "style: go_zero\n" {
				t.Fatalf("%s config does not use the default style: %q", frame, rendered)
			}
		})
	}
}

func TestAPIRPCConfigTemplatesPreserveExplicitStyle(t *testing.T) {
	for _, frame := range []string{"api", "rpc"} {
		t.Run(frame, func(t *testing.T) {
			path := ".template/frame/" + frame + "/app/.pzero.yaml.tpl"
			content := readEmbeddedTemplate(t, path)
			rendered, err := pzerotemplatex.ParseTemplate(path, map[string]any{
				"Style": "gozero",
			}, []byte(content))
			if err != nil {
				t.Fatalf("render %s config template: %v", frame, err)
			}
			if string(rendered) != "style: gozero\n" {
				t.Fatalf("%s config does not preserve explicit style: %q", frame, rendered)
			}
		})
	}
}
