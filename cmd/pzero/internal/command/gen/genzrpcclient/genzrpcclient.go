package genzrpcclient

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pkg/errors"
	"github.com/rinchsan/gosimports"
	"github.com/samber/lo"
	conf "github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/execx"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/generator"
	rpcparser "github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	"github.com/polpo-space/pzero/cmd/pzero/internal/desc"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console/progress"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/mod"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/osx"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/templatex"
)

type DirContext struct {
	ImportBase      string
	PbPackage       string
	OptionGoPackage string
	Resource        string
	Output          string
}

func (d DirContext) GetCall() generator.Dir {
	fileName := filepath.Join(d.Output, "typed", d.Resource)
	return generator.Dir{
		Filename: fileName,
		GetChildPackage: func(childPath string) (string, error) {
			return strings.ToLower(childPath), nil
		},
	}
}

func (d DirContext) GetEtc() generator.Dir {
	panic("implement me")
}

func (d DirContext) GetInternal() generator.Dir {
	panic("implement me")
}

func (d DirContext) GetConfig() generator.Dir {
	panic("implement me")
}

func (d DirContext) GetLogic() generator.Dir {
	panic("implement me")
}

func (d DirContext) GetServer() generator.Dir {
	panic("implement me")
}

func (d DirContext) GetSvc() generator.Dir {
	panic("implement me")
}

func (d DirContext) GetPb() generator.Dir {
	return generator.Dir{
		Package: d.packagePath(),
	}
}

func (d DirContext) packagePath() string {
	packagePath := filepath.ToSlash(fmt.Sprintf("%s/model%s/%s", d.ImportBase, d.Resource, strings.TrimPrefix(d.OptionGoPackage, "./")))
	return packagePath
}

func (d DirContext) GetProtoGo() generator.Dir {
	return generator.Dir{
		Filename: d.OptionGoPackage,
		Package:  d.packagePath(),
	}
}

func (d DirContext) GetMain() generator.Dir {
	panic("implement me")
}

func (d DirContext) GetServiceName() stringx.String {
	panic("implement me")
}

func (d DirContext) SetPbDir(pbDir, grpcDir string) {
	panic("implement me")
}

func Generate(genModule bool) (err error) {
	showProgress := !config.C.Quiet
	files, err := listZRPCClientProtoFiles()
	if err != nil {
		return err
	}

	if len(files) > 0 {
		if err = executeStage(
			console.Green("Gen")+" "+console.Yellow("zrpcclient"),
			showProgress,
			showProgress,
			func(progressChan chan<- progress.Message) error {
				return generateMainZRPCClient(genModule, files, progressChan)
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func HasInput() (bool, error) {
	files, err := listZRPCClientProtoFiles()
	if err != nil {
		return false, err
	}
	return len(files) > 0, nil
}

func executeStage(title string, headerShown, showProgress bool, fn func(chan<- progress.Message) error) error {
	if !showProgress {
		return fn(nil)
	}

	progressChan := make(chan progress.Message, 10)
	done := make(chan struct{})
	var stageErr error

	if headerShown {
		fmt.Printf("%s\n", console.BoxHeader("", title))
	}

	go func() {
		stageErr = fn(progressChan)
		close(done)
	}()

	state := progress.ConsumeStage(progressChan, done, title, false, headerShown)
	progress.FinishStage(title, false, &state, stageErr)

	if stageErr != nil {
		return console.MarkRenderedError(stageErr)
	}

	return nil
}

func listZRPCClientProtoFiles() ([]string, error) {
	var files []string

	switch {
	case len(config.C.Gen.Zrpcclient.Desc) > 0:
		for _, v := range config.C.Gen.Zrpcclient.Desc {
			if !osx.IsDir(v) {
				if filepath.Ext(v) == ".proto" {
					files = append(files, v)
				}
				continue
			}

			specifiedProtoFiles, err := desc.FindRpcServiceProtoFiles(v)
			if err != nil {
				return nil, err
			}
			files = append(files, specifiedProtoFiles...)
		}
	default:
		if pathx.FileExists(config.C.ProtoDir()) {
			var err error
			files, err = desc.FindRpcServiceProtoFiles(config.C.ProtoDir())
			if err != nil {
				return nil, err
			}
		}
	}

	for _, v := range config.C.Gen.Zrpcclient.DescIgnore {
		if !osx.IsDir(v) {
			if filepath.Ext(v) == ".proto" {
				files = lo.Reject(files, func(item string, _ int) bool {
					return item == v
				})
			}
			continue
		}

		specifiedProtoFiles, err := desc.FindRpcServiceProtoFiles(v)
		if err != nil {
			return nil, err
		}
		for _, saf := range specifiedProtoFiles {
			files = lo.Reject(files, func(item string, _ int) bool {
				return item == saf
			})
		}
	}

	return files, nil
}

func generateMainZRPCClient(genModule bool, files []string, progressChan chan<- progress.Message) error {
	g := generator.NewGenerator(config.C.Style, false)

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	excludeThirdPartyProtoFiles, err := desc.FindExcludeThirdPartyProtoFiles(config.C.ProtoDir())
	if err != nil {
		return err
	}

	var services []string
	for _, fp := range files {
		parser := rpcparser.NewDefaultProtoParser()
		parse, err := parser.Parse(fp, true)
		if err != nil {
			return err
		}
		dirContext := DirContext{
			ImportBase:      filepath.Join(config.C.Gen.Zrpcclient.GoModule),
			PbPackage:       parse.PbPackage,
			OptionGoPackage: parse.GoPackage,
			Output:          config.C.Gen.Zrpcclient.Output,
		}
		for _, service := range parse.Service {
			services = append(services, service.Name)
			_ = os.MkdirAll(filepath.Join(dirContext.GetCall().Filename, strings.ToLower(service.Name)), 0o755)
		}

		pbDir := filepath.Join(config.C.Gen.Zrpcclient.Output, "model")
		if err = os.MkdirAll(pbDir, 0o755); err != nil {
			return err
		}

		var importPaths []string
		importPaths = append(importPaths, config.C.ProtoDir())
		importPaths = append(importPaths, filepath.Join(config.C.ProtoDir(), "third_party"))

		var protoParser protoparse.Parser
		protoParser.InferImportPaths = false

		protoDir := filepath.Join("desc", "proto")
		thirdPartyProtoDir := filepath.Join("desc", "proto", "third_party")
		protoParser.ImportPaths = []string{protoDir, thirdPartyProtoDir}
		for _, v := range config.C.Gen.Zrpcclient.ProtoInclude {
			protoParser.ImportPaths = append(protoParser.ImportPaths, v)
		}
		protoParser.IncludeSourceCodeInfo = true

		rel, err := filepath.Rel(config.C.ProtoDir(), fp)
		if err != nil {
			return err
		}

		fds, err := protoParser.ParseFiles(rel)
		if err != nil {
			return err
		}

		if len(fds) == 0 {
			continue
		}

		goPackage := fds[0].AsFileDescriptorProto().GetOptions().GetGoPackage()

		getMod, err := mod.GetGoMod(config.C.Wd())
		if err != nil {
			return err
		}

		module := config.C.Gen.Zrpcclient.GoModule
		if !genModule && config.C.Gen.Zrpcclient.Output != "." {
			module = getMod.Path
		}

		protocCmd := fmt.Sprintf("protoc %s -I%s -I%s --go_out=%s --go-grpc_out=%s",
			fp,
			config.C.ProtoDir(),
			filepath.Join(config.C.ProtoDir(), "third_party"),
			func() string {
				if !genModule {
					return "."
				}
				return filepath.Join(config.C.Gen.Zrpcclient.Output)
			}(),
			func() string {
				if !genModule {
					return "."
				}
				return filepath.Join(config.C.Gen.Zrpcclient.Output)
			}(),
		)

		for _, exp := range excludeThirdPartyProtoFiles {
			rel, err = filepath.Rel(config.C.ProtoDir(), exp)
			if err != nil {
				return err
			}

			fds, err = protoParser.ParseFiles(rel)
			if err != nil {
				return err
			}

			if len(fds) == 0 {
				continue
			}

			goPackage = fds[0].AsFileDescriptorProto().GetOptions().GetGoPackage()

			protocCmd += fmt.Sprintf(" --go_opt=module=%s --go_opt=M%s=%s --go-grpc_opt=module=%s --go-grpc_opt=M%s=%s", module, rel, func() string {
				if strings.HasPrefix(goPackage, module) {
					return goPackage
				}
				if genModule {
					return filepath.ToSlash(filepath.Join(module, "model", goPackage))
				}
				return filepath.ToSlash(filepath.Join(module, config.C.Gen.Zrpcclient.Output, "model", goPackage))
			}(), module, rel, func() string {
				if strings.HasPrefix(goPackage, module) {
					return goPackage
				}
				if genModule {
					return filepath.ToSlash(filepath.Join(module, "model", goPackage))
				}
				return filepath.ToSlash(filepath.Join(module, config.C.Gen.Zrpcclient.Output, "model", goPackage))
			}())
		}

		if len(config.C.Gen.Zrpcclient.ProtoInclude) > 0 {
			protocCmd += fmt.Sprintf(" -I%s ", strings.Join(config.C.Gen.Zrpcclient.ProtoInclude, " -I"))
		}

		resp, err := execx.Run(protocCmd, wd)
		if err != nil {
			return errors.Errorf("err: [%v], resp: [%s]", err, resp)
		}

		err = g.GenCall(dirContext, parse, &conf.Config{
			NamingFormat: config.C.Style,
		}, &generator.ZRpcContext{
			Multiple:    true,
			IsGenClient: true,
		})
		if err != nil {
			return err
		}

		if progressChan != nil {
			progressChan <- progress.NewFile(fp)
		}
	}

	template, err := templatex.ParseTemplate(filepath.ToSlash(filepath.Join("client", "zrpcclient-go", "clientset.go.tpl")), map[string]any{
		"Module":   config.C.Gen.Zrpcclient.GoModule,
		"Package":  config.C.Gen.Zrpcclient.GoPackage,
		"Services": services,
	}, embeded.ReadTemplateFile(filepath.ToSlash(filepath.Join("client", "zrpcclient-go", "clientset.go.tpl"))))
	if err != nil {
		return err
	}

	formated, err := gosimports.Process("", template, nil)
	if err != nil {
		return errors.Errorf("format go file %s %s meet error: %v", filepath.Join(config.C.Gen.Zrpcclient.Output, "clientset.go"), template, err)
	}
	if err = os.WriteFile(filepath.Join(config.C.Gen.Zrpcclient.Output, "clientset.go"), formated, 0o644); err != nil {
		return err
	}

	if genModule {
		goVersion, err := mod.GetGoVersion()
		if err != nil {
			return err
		}
		templateData := map[string]any{
			"GoVersion": goVersion,
			"GoArch":    runtime.GOARCH,
		}
		templateData["Module"] = config.C.Gen.Zrpcclient.GoModule
		if config.C.Gen.Zrpcclient.GoVersion != "" {
			templateData["GoVersion"] = config.C.Gen.Zrpcclient.GoVersion
		}
		template, err = templatex.ParseTemplate(filepath.ToSlash(filepath.Join("client", "zrpcclient-go", "go.mod.tpl")), templateData, embeded.ReadTemplateFile(filepath.ToSlash(filepath.Join("client", "zrpcclient-go", "go.mod.tpl"))))
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(config.C.Gen.Zrpcclient.Output, "go.mod"), template, 0o644); err != nil {
			return err
		}
	}

	return genNoRpcServiceExcludeThirdPartyProto(genModule, config.C.Gen.Zrpcclient.GoModule, progressChan)
}

func genNoRpcServiceExcludeThirdPartyProto(genModule bool, module string, progressChan chan<- progress.Message) error {
	excludeThirdPartyProtoFiles, err := desc.FindNoRpcServiceExcludeThirdPartyProtoFiles(config.C.ProtoDir())
	if err != nil {
		return err
	}

	var protoParser protoparse.Parser
	protoParser.InferImportPaths = false

	protoDir := filepath.Join("desc", "proto")
	thirdPartyProtoDir := filepath.Join("desc", "proto", "third_party")
	protoParser.ImportPaths = []string{protoDir, thirdPartyProtoDir}
	protoParser.IncludeSourceCodeInfo = true

	pbDir := filepath.Join(config.C.Gen.Zrpcclient.Output, "model")
	err = os.MkdirAll(pbDir, 0o755)
	if err != nil {
		return err
	}

	for _, v := range excludeThirdPartyProtoFiles {
		rel, err := filepath.Rel(config.C.ProtoDir(), v)
		if err != nil {
			return err
		}

		fds, err := protoParser.ParseFiles(rel)
		if err != nil {
			return err
		}

		if len(fds) == 0 {
			continue
		}

		goPackage := fds[0].AsFileDescriptorProto().GetOptions().GetGoPackage()

		getMod, err := mod.GetGoMod(config.C.Wd())
		if err != nil {
			return err
		}

		if !genModule {
			if config.C.Gen.Zrpcclient.Output != "." {
				module = getMod.Path
			}
		}

		command := fmt.Sprintf("protoc %s -I%s -I%s --go_out=%s --go_opt=module=%s --go_opt=M%s=%s --go-grpc_out=%s --go-grpc_opt=module=%s --go-grpc_opt=M%s=%s",
			v,
			config.C.ProtoDir(),
			filepath.Join(config.C.ProtoDir(), "third_party"),
			func() string {
				if !genModule {
					return "."
				}
				return filepath.Join(config.C.Gen.Zrpcclient.Output)
			}(),
			module,
			rel,
			func() string {
				if strings.HasPrefix(goPackage, module) {
					return goPackage
				}
				if genModule {
					return filepath.ToSlash(filepath.Join(module, "model", goPackage))
				}
				return filepath.ToSlash(filepath.Join(module, config.C.Gen.Zrpcclient.Output, "model", goPackage))
			}(),
			func() string {
				if !genModule {
					return "."
				}
				return filepath.Join(config.C.Gen.Zrpcclient.Output)
			}(),
			module,
			rel,
			func() string {
				if strings.HasPrefix(goPackage, module) {
					return goPackage
				}

				if genModule {
					return filepath.ToSlash(filepath.Join(module, "model", goPackage))
				}
				return filepath.ToSlash(filepath.Join(module, config.C.Gen.Zrpcclient.Output, "model", goPackage))
			}(),
		)

		// Debug removed(command)

		_, err = execx.Run(command, config.C.Wd())
		if err != nil {
			return err
		}
		if progressChan != nil {
			progressChan <- progress.NewFile(v)
		}
	}
	return nil
}

