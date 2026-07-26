package templatex

import (
	"strings"

	"github.com/polpo-space/pzero/core/templatex"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
)

// ParseTemplate template
func ParseTemplate(name string, data map[string]any, tplT []byte) ([]byte, error) {
	for _, v := range config.C.RegisterTplVal {
		split := strings.Split(v, "=")
		if len(split) == 2 {
			data[split[0]] = split[1]
		}
	}
	return templatex.ParseTemplateWithName(name, data, tplT, templatex.WithFuncMaps(registerFuncMap))
}
