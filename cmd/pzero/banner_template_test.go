package main

import (
	"strings"
	"testing"
)

func TestBannerTemplatesUseServiceName(t *testing.T) {
	tests := []struct {
		name       string
		frame      string
		serviceRef string
	}{
		{name: "api", frame: "api", serviceRef: "c.Rest.Name"},
		{name: "rpc", frame: "rpc", serviceRef: "c.Zrpc.Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configTemplate := readEmbeddedTemplate(t,
				".template/frame/"+tt.frame+"/app/internal/config/config.go.tpl")
			if strings.Contains(configTemplate, "BannerConf") {
				t.Fatalf("%s config template still declares BannerConf", tt.frame)
			}

			serverTemplate := readEmbeddedTemplate(t,
				".template/frame/"+tt.frame+"/app/cmd/server.go.tpl")
			if !strings.Contains(serverTemplate, "printBanner("+tt.serviceRef+")") {
				t.Fatalf("%s server template does not use %s", tt.frame, tt.serviceRef)
			}
			if !strings.Contains(serverTemplate,
				`figure.NewColorFigure(serviceName, "starwars", "green", false).Print()`) {
				t.Fatalf("%s server template does not render the service name safely", tt.frame)
			}
		})
	}
}

func readEmbeddedTemplate(t *testing.T, path string) string {
	t.Helper()

	content, err := Template.ReadFile(path)
	if err != nil {
		t.Fatalf("read embedded template %s: %v", path, err)
	}

	return string(content)
}
