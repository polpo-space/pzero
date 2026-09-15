package genmodel

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rinchsan/gosimports"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/templatex"
)

type TableInfo struct {
	Name              string
	HasCacheExpiry    bool
	HasNotFoundExpiry bool
}

func (jm *PzeroModel) GenRegister(tables []string) error {
	// Debug removed("get register tables: %v", tables)

	slices.Sort(tables)

	var imports []string
	var tablePackages []string
	var tableInfos []TableInfo

	for _, table := range tables {
		name := strings.ToLower(table)
		if !pathx.FileExists(filepath.Join("internal", "model", name)) {
			continue
		}
		imports = append(imports, fmt.Sprintf("%s/internal/model/%s", jm.Module, name))
		tablePackages = append(tablePackages, name)
		tableInfos = append(tableInfos, TableInfo{Name: name})
	}

	// Build cache expiry table maps - only when cache is enabled
	var modelExpiryTable map[string]int64
	var modelNotFoundExpiryTable map[string]int64
	if config.C.Gen.ModelCache {
		modelExpiryTable = make(map[string]int64)
		modelNotFoundExpiryTable = make(map[string]int64)
		for _, v := range config.C.Gen.ModelCacheExpiryTable {
			if v.Expiry > 0 {
				modelExpiryTable[v.Table] = v.Expiry
			}
			if v.NotFoundExpiry > 0 {
				modelNotFoundExpiryTable[v.Table] = v.NotFoundExpiry
			}
		}
	}

	// Update tableInfos with cache expiry info
	for i := range tableInfos {
		if modelExpiryTable != nil {
			if _, ok := modelExpiryTable[tableInfos[i].Name]; ok {
				tableInfos[i].HasCacheExpiry = true
			}
		}
		if modelNotFoundExpiryTable != nil {
			if _, ok := modelNotFoundExpiryTable[tableInfos[i].Name]; ok {
				tableInfos[i].HasNotFoundExpiry = true
			}
		}
	}

	template, err := templatex.ParseTemplate(filepath.Join("model", "model.go.tpl"), map[string]any{
		"Imports":                  imports,
		"TablePackages":            tablePackages,
		"TableInfos":               tableInfos,
		"ModelExpiryTable":         modelExpiryTable,
		"ModelNotFoundExpiryTable": modelNotFoundExpiryTable,
		"ModelCache":               config.C.Gen.ModelCache,
	}, embeded.ReadTemplateFile(filepath.Join("model", "model.go.tpl")))
	if err != nil {
		return err
	}

	format, err := gosimports.Process("", template, nil)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join("internal", "model", "model.go"), format, 0o644)
}
