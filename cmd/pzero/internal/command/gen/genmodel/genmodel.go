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
	pzerodesc "github.com/polpo-space/pzero/cmd/pzero/internal/desc"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console/progress"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/dsn"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/filex"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/osx"
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

	if !pathx.FileExists(config.C.SqlDir()) && !config.C.Gen.ModelDatasource {
		return nil, nil
	}

	driver, err := normalizeModelDriver(config.C.Gen.ModelDriver)
	if err != nil {
		return nil, err
	}
	config.C.Gen.ModelDriver = driver

	if !config.C.Gen.ModelDatasource {
		return nil, errors.New("postgres model only supports datasource mode")
	}
	if hasSQLInput() {
		return nil, errors.New("postgres model generation only supports datasource input")
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
		err := conn.SqlConn.QueryRowsCtx(ctx, &tables, fmt.Sprintf("select tablename from pg_tables where schemaname = '%s'", config.C.Gen.ModelSchema))
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

func hasSQLInput() bool {
	if len(config.C.Gen.Desc) != 0 {
		for _, v := range config.C.Gen.Desc {
			if filepath.Ext(v) == ".sql" || osx.IsDir(v) {
				return true
			}
		}
	}

	if pathx.FileExists(config.C.SqlDir()) {
		sqlFiles, err := pzerodesc.FindSqlFiles(config.C.SqlDir())
		return err == nil && len(sqlFiles) > 0
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
