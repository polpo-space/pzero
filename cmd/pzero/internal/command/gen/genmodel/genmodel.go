package genmodel

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/postgres"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"golang.org/x/sync/errgroup"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console/progress"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/dsn"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/filex"
)

type PzeroModel struct {
	Module string
}

type Conn struct {
	Schema  string
	SqlConn sqlx.SqlConn
}

func (jm *PzeroModel) Gen(progressChan chan<- progress.Message) ([]string, error) {
	var (
		allTables []string
		err       error
		conns     []Conn
		genFiles  []string
	)

	// Model gen 仅由 model-datasource 开关驱动。
	// desc/sql 是 schema snapshot，可与 datasource 共存，不再作为 gen 入口或失败条件。
	if !config.C.Gen.ModelDatasource {
		return nil, nil
	}

	driver, err := normalizeModelDriver(config.C.Gen.ModelDriver)
	if err != nil {
		return nil, err
	}
	config.C.Gen.ModelDriver = driver

	if hasExplicitSQLDesc() {
		return nil, errors.New("postgres model generation only supports datasource input; desc/sql is a schema snapshot and is not accepted via --desc")
	}
	if len(config.C.Gen.ModelDatasourceUrl) == 0 {
		return nil, errors.New("model-datasource-url is required when model-datasource is enabled")
	}

	for _, v := range config.C.Gen.ModelDatasourceUrl {
		meta, err := dsn.ParseDSN(config.C.Gen.ModelDriver, v)
		if err != nil {
			return nil, err
		}
		conns = append(conns, Conn{
			Schema:  meta[dsn.Database],
			SqlConn: postgres.New(v),
		})
	}

	// 处理模板
	var goctlHome string
	tempDir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// 先写入内置模板
	err = embeded.WriteTemplateDir(filepath.Join("go-zero", "model"), filepath.Join(tempDir, "model"))
	if err != nil {
		return nil, err
	}

	// 如果用户自定义了模板，则复制覆盖
	customTemplatePath := filepath.Join(config.C.Home, "go-zero", "model")
	if pathx.FileExists(customTemplatePath) {
		err = filex.CopyDir(customTemplatePath, filepath.Join(tempDir, "model"))
		if err != nil {
			return nil, err
		}
	}

	goctlHome = tempDir

	// --desc / gen.desc 用于限定 api/proto 生成范围；有 desc 时跳过 model，避免无意义全表 regen。
	if len(config.C.Gen.Desc) != 0 {
		return nil, nil
	}

	if len(config.C.Gen.ModelDatasourceTable) == 1 && config.C.Gen.ModelDatasourceTable[0] == "*" {
		allTables, err = getAllTables(conns)
		if err != nil {
			return nil, err
		}
	} else {
		allTables = config.C.Gen.ModelDatasourceTable
	}

	// For datasource mode, generate code for each table directly
	var eg errgroup.Group
	eg.SetLimit(len(allTables))
	for _, tableName := range allTables {
		eg.Go(func() error {
			if progressChan != nil {
				progressChan <- progress.NewFile(tableName)
			}
			return generateModelFromDatasource(tableName, goctlHome)
		})
	}
	if err = eg.Wait(); err != nil {
		return nil, err
	}
	err = jm.GenRegister(allTables)
	if err != nil {
		return nil, err
	}
	// Add table names to genFiles for datasource mode
	genFiles = append(genFiles, allTables...)
	return genFiles, nil
}

func getAllTables(conns []Conn) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var allTables []string

	if config.C.Gen.ModelSchema == "" {
		config.C.Gen.ModelSchema = "public"
	}
	for _, conn := range conns {
		var tables []string
		err := conn.SqlConn.QueryRowsCtx(ctx, &tables, "select tablename from pg_tables where schemaname = $1", config.C.Gen.ModelSchema)
		if err != nil {
			return nil, err
		}
		for _, v := range tables {
			allTables = append(allTables, v)
		}
	}
	return allTables, nil
}

func normalizeModelDriver(driver string) (string, error) {
	switch driver {
	case "", "postgres", "pgx":
		return "pgx", nil
	default:
		return "", errors.Errorf("model driver %s not support, only postgres is supported", driver)
	}
}

// hasExplicitSQLDesc reports whether gen.desc/--desc explicitly targets SQL snapshot
// paths. Presence of desc/sql alone is fine and must not block datasource model gen.
func hasExplicitSQLDesc() bool {
	if len(config.C.Gen.Desc) == 0 {
		return false
	}

	sqlDir := filepath.Clean(config.C.SqlDir())
	sqlPrefix := sqlDir + string(os.PathSeparator)

	for _, v := range config.C.Gen.Desc {
		clean := filepath.Clean(v)
		if filepath.Ext(clean) == ".sql" {
			return true
		}
		if clean == sqlDir || strings.HasPrefix(clean, sqlPrefix) {
			return true
		}
	}
	return false
}

func getIsCacheTable(t string) bool {
	if config.C.Gen.ModelCache && len(config.C.Gen.ModelCacheTable) == 1 && config.C.Gen.ModelCacheTable[0] == "*" {
		return true
	}

	if config.C.Gen.ModelCache {
		for _, v := range config.C.Gen.ModelCacheTable {
			if v == t {
				return true
			}
		}
	}
	return false
}

func getIgnoreColumns(tableName string) []string {
	if config.C.Gen.ModelIgnoreColumnsTable != nil {
		for _, v := range config.C.Gen.ModelIgnoreColumnsTable {
			if v.Table == tableName {
				return v.Columns
			}
		}
	}
	return config.C.Gen.ModelIgnoreColumns
}

func generateModelFromDatasource(tableName, goctlHome string) error {
	bf := tableName
	if strings.Contains(tableName, ".") {
		bf = strings.Split(tableName, ".")[1]
	}

	var (
		modelDir string
		schema   = config.C.Gen.ModelSchema
	)

	if strings.Contains(tableName, ".") {
		split := strings.Split(tableName, ".")
		modelDir = filepath.Join("internal", "model", split[0], strings.ToLower(split[1]))
	} else {
		modelDir = filepath.Join("internal", "model", strings.ToLower(bf))
	}

	if schema == "" {
		schema = "public"
	}
	var datasourceUrl string
	if strings.Contains(tableName, ".") {
		for _, v := range config.C.Gen.ModelDatasourceUrl {
			meta, err := dsn.ParseDSN("pgx", v)
			if err != nil {
				return err
			}
			if meta[dsn.Database] == strings.Split(tableName, ".")[0] {
				datasourceUrl = v
				break
			}
		}
	} else {
		datasourceUrl = config.C.Gen.ModelDatasourceUrl[0]
	}

	cmd := exec.Command("goctl", "model", "pg", "datasource", "--url", datasourceUrl, "--schema", schema, "-t", bf, "--dir", modelDir, "--home", goctlHome, "--style", config.C.Style, "-i", strings.Join(getIgnoreColumns(bf), ","), "--cache="+fmt.Sprintf("%t", getIsCacheTable(bf)), "-p", config.C.Gen.ModelCachePrefix, "--strict="+fmt.Sprintf("%t", config.C.Gen.ModelStrict))
	// Debug removed(cmd.String())
	resp, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Errorf("gen model code meet error. Err: %s:%s", err.Error(), resp)
	}

	return nil
}
