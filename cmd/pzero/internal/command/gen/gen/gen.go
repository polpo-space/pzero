package gen

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rinchsan/gosimports"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"

	"github.com/polpo-space/pzero/cmd/pzero/internal/command/gen/genapi"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/gen/genmodel"
	"github.com/polpo-space/pzero/cmd/pzero/internal/command/gen/genrpc"
	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/desc"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console/progress"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/mod"
)

func Run() error {
	module, err := resolveModule()
	if err != nil {
		return err
	}

	defer func() {
		RemoveExtraFiles(config.C.Wd(), config.C.Style)
	}()

	// --desc 限定 api/proto 时跳过 model；指向 sql 仍进入 model 段，由 genmodel 报错。
	if shouldRunModelStage() {
		if err := runModelStage(module); err != nil {
			return err
		}
	}

	pzeroApi := genapi.PzeroApi{
		Module: module,
	}

	apiTitle := console.Green("Gen") + " " + console.Yellow("api")

	apiHeaderShown := !config.C.Quiet && pathx.FileExists(config.C.ApiDir())

	// Generate api
	apiProgressChan := make(chan progress.Message, 10)
	apiDone := make(chan struct{})
	var apiErr error

	// Show box header immediately before starting goroutine
	if apiHeaderShown {
		fmt.Printf("%s\n", console.BoxHeader("", apiTitle))
	}

	go func() {
		apiErr = pzeroApi.Gen(apiProgressChan)
		close(apiDone)
	}()

	apiState := progress.ConsumeStage(apiProgressChan, apiDone, apiTitle, config.C.Quiet, apiHeaderShown)
	progress.FinishStage(apiTitle, config.C.Quiet, &apiState, apiErr)

	if apiErr != nil {
		return apiErr
	}

	pzeroRpc := genrpc.PzeroRpc{
		Module: module,
	}

	rpcTitle := console.Green("Gen") + " " + console.Yellow("rpc")

	// Generate rpc
	rpcProgressChan := make(chan progress.Message, 10)
	rpcDone := make(chan struct{})
	var rpcErr error
	go func() {
		rpcErr = pzeroRpc.Gen(rpcProgressChan)
		close(rpcProgressChan)
		close(rpcDone)
	}()

	rpcState := progress.ConsumeStage(rpcProgressChan, rpcDone, rpcTitle, config.C.Quiet, false)
	progress.FinishStage(rpcTitle, config.C.Quiet, &rpcState, rpcErr)

	if rpcErr != nil {
		return rpcErr
	}

	return nil
}

func RunModel() error {
	module, err := resolveModule()
	if err != nil {
		return err
	}
	return runModelStage(module)
}

func resolveModule() (string, error) {
	moduleStruct, err := mod.GetGoMod(config.C.Wd())
	if err != nil {
		return "", errors.Wrapf(err, "get go module struct error")
	}
	module := moduleStruct.Path
	gosimports.LocalPrefix = module

	if !pathx.FileExists("go.mod") {
		module, err = mod.GetParentPackage(config.C.Wd())
		if err != nil {
			return "", errors.Wrapf(err, "get parent package error")
		}
	}
	return module, nil
}

func shouldRunModelStage() bool {
	return len(config.C.Gen.Desc) == 0 || genmodel.HasExplicitSQLDesc()
}

func runModelStage(module string) error {
	pzeroModel := genmodel.PzeroModel{
		Module: module,
	}

	modelTitle := console.Green("Gen") + " " + console.Yellow("model")

	progressChan := make(chan progress.Message, 10)
	done := make(chan struct{})
	var modelErr error
	go func() {
		_, modelErr = pzeroModel.Gen(progressChan)
		close(done)
	}()

	modelState := progress.ConsumeStage(progressChan, done, modelTitle, config.C.Quiet, false)
	progress.FinishStage(modelTitle, config.C.Quiet, &modelState, modelErr)
	return modelErr
}

func RemoveExtraFiles(wd, style string) {
	if pathx.FileExists(filepath.Join("desc", "api")) {
		apiFilenames, err := desc.FindApiFiles(filepath.Join("desc", "api"))
		if err == nil {
			for _, v := range apiFilenames {
				if desc.GetApiFrameMainGoFilename(wd, v, style) != "main.go" {
					_ = os.Remove(filepath.Join(wd, desc.GetApiFrameMainGoFilename(wd, v, style)))
				}
				if desc.GetApiFrameEtcFilename(wd, v, style) != "etc.yaml" {
					_ = os.Remove(filepath.Join(wd, "etc", desc.GetApiFrameEtcFilename(wd, v, style)))
				}
			}
		}
	}

	if pathx.FileExists(filepath.Join("desc", "proto")) {
		protoFilenames, err := desc.FindRpcServiceProtoFiles(filepath.Join("desc", "proto"))
		if err == nil {
			for _, v := range protoFilenames {
				v = filepath.Base(v)
				fileBase := v[0 : len(v)-len(path.Ext(v))]
				if desc.GetProtoFrameMainGoFilename(fileBase, style) != "main.go" {
					_ = os.Remove(filepath.Join(wd, desc.GetProtoFrameMainGoFilename(fileBase, style)))
				}
				if desc.GetProtoFrameEtcFilename(fileBase, style) != "etc.yaml" {
					_ = os.Remove(filepath.Join(wd, "etc", desc.GetProtoFrameEtcFilename(fileBase, style)))
				}
			}
		}
	}

	// Check if etc directory is empty and remove it if so
	etcDir := filepath.Join(wd, "etc")
	if pathx.FileExists(etcDir) {
		entries, err := os.ReadDir(etcDir)
		if err == nil && len(entries) == 0 {
			if err := os.Remove(etcDir); err != nil && !errors.Is(err, os.ErrNotExist) {
			}
		}
	}
}
