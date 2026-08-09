package genrpc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/samber/lo"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/execx"
	rpcparser "github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	goctlconsole "github.com/zeromicro/go-zero/tools/goctl/util/console"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	pzerodesc "github.com/polpo-space/pzero/cmd/pzero/internal/desc"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console/progress"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/filex"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/gitstatus"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/osx"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/stringx"
)

type PzeroRpc struct {
	Module string
}

type (
	ImportLines []string

	RegisterLines []string
)

func (l ImportLines) String() string {
	return "\n\n\t" + strings.Join(l, "\n\t")
}

func (l RegisterLines) String() string {
	return "\n\t\t" + strings.Join(l, "\n\t\t")
}

func (jr *PzeroRpc) Gen(progressChan chan<- progress.Message) (map[string]*rpcparser.Proto, error) {
	var (
		serverImports   ImportLines
		pbImports       ImportLines
		registerServers RegisterLines
	)

	protoDirs := config.C.ProtoDirs()

	// 获取全量 proto 文件（支持多 proto-dir）
	protoFiles, err := findRpcServiceProtoFilesInDirs(protoDirs)
	if err != nil {
		return nil, err
	}

	if len(protoFiles) == 0 {
		return nil, nil
	}

	protoSpecMap := make(map[string]*rpcparser.Proto, len(protoFiles))
	for _, v := range protoFiles {
		protoParser := rpcparser.NewDefaultProtoParser()
		var parse rpcparser.Proto
		parse, err = protoParser.Parse(v, true)
		if err != nil {
			return nil, err
		}
		protoSpecMap[v] = &parse
	}

	// 获取需要生成代码的 proto 文件
	var genCodeProtoFiles []string
	genCodeProtoSpecMap := make(map[string]*rpcparser.Proto, len(protoFiles))

	switch {
	case config.C.Gen.GitChange && gitstatus.IsGitRepo(filepath.Join(config.C.Wd())) && len(config.C.Gen.Desc) == 0:
		for _, dir := range protoDirs {
			m, _, err := gitstatus.ChangedFiles(dir, ".proto")
			if err != nil {
				continue
			}
			genCodeProtoFiles = append(genCodeProtoFiles, m...)
			for _, file := range m {
				genCodeProtoSpecMap[file] = protoSpecMap[file]
			}
		}
	case len(config.C.Gen.Desc) > 0:
		for _, v := range config.C.Gen.Desc {
			if !osx.IsDir(v) {
				if filepath.Ext(v) == ".proto" {
					cleaned := filepath.Clean(v)
					genCodeProtoFiles = append(genCodeProtoFiles, filepath.Join(strings.Split(filepath.ToSlash(v), "/")...))
					genCodeProtoSpecMap[cleaned] = protoSpecMap[cleaned]
				}
			} else {
				specifiedProtoFiles, err := pzerodesc.FindRpcServiceProtoFiles(v)
				if err != nil {
					return nil, err
				}
				genCodeProtoFiles = append(genCodeProtoFiles, specifiedProtoFiles...)
				for _, saf := range specifiedProtoFiles {
					genCodeProtoSpecMap[filepath.Clean(saf)] = protoSpecMap[filepath.Clean(saf)]
				}
			}
		}
	default:
		genCodeProtoFiles = protoFiles
		genCodeProtoSpecMap = protoSpecMap
	}

	// ignore proto desc
	for _, v := range config.C.Gen.DescIgnore {
		if !osx.IsDir(v) {
			if filepath.Ext(v) == ".proto" {
				genCodeProtoFiles = lo.Reject(genCodeProtoFiles, func(item string, _ int) bool {
					return item == filepath.Clean(v)
				})
				protoFiles = lo.Reject(protoFiles, func(item string, _ int) bool {
					return item == filepath.Clean(v)
				})
				delete(genCodeProtoSpecMap, filepath.Clean(v))
				delete(protoSpecMap, filepath.Clean(v))
			}
		} else {
			specifiedProtoFiles, err := pzerodesc.FindRpcServiceProtoFiles(v)
			if err != nil {
				return nil, err
			}
			for _, saf := range specifiedProtoFiles {
				genCodeProtoFiles = lo.Reject(genCodeProtoFiles, func(item string, _ int) bool {
					return item == saf
				})
				protoFiles = lo.Reject(protoFiles, func(item string, _ int) bool {
					return item == saf
				})
				delete(genCodeProtoSpecMap, saf)
				delete(protoSpecMap, saf)
			}
		}
	}

	if len(genCodeProtoFiles) == 0 {
		return protoSpecMap, nil
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), "pzero-rpc-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	customTemplatePath := filepath.Join(config.C.Home, "go-zero", "rpc")
	if pathx.FileExists(customTemplatePath) {
		err = filex.CopyDir(customTemplatePath, filepath.Join(tempDir, "rpc"))
		if err != nil {
			return nil, err
		}
	}
	goctlHome := tempDir

	pbOutDirExternal := filepath.Join(tempDir, "pbout")
	if err = os.MkdirAll(pbOutDirExternal, 0o755); err != nil {
		return nil, err
	}

	excludeThirdPartyProtoFiles, err := findExcludeThirdPartyProtoFilesInDirs(protoDirs)
	if err != nil {
		return nil, err
	}

	importPaths := buildProtoImportPaths(protoDirs)
	var protoParser protoparse.Parser
	protoParser.InferImportPaths = false
	protoParser.ImportPaths = importPaths
	protoParser.IncludeSourceCodeInfo = true

	for _, v := range protoFiles {
		allLogicFiles, err := jr.GetAllLogicFiles(v, protoSpecMap[v])
		if err != nil {
			return nil, err
		}

		allServerFiles, err := jr.GetAllServerFiles(v, protoSpecMap[v])
		if err != nil {
			return nil, err
		}

		fileExternal := isExternalGoPackage(protoSpecMap[v].GoPackage)

		if lo.Contains(genCodeProtoFiles, v) {
			protoRoot, rel, err := relToProtoDir(v, protoDirs)
			if err != nil {
				return nil, err
			}

			fds, err := protoParser.ParseFiles(rel)
			if err != nil {
				return nil, err
			}
			if len(fds) == 0 {
				continue
			}

			pbOutDir := "."
			if fileExternal {
				pbOutDir = pbOutDirExternal
			}

			includeArgs := buildProtocIncludeArgs(protoDirs)
			command := fmt.Sprintf("goctl rpc protoc %s%s --go_out=%s --go-grpc_out=%s --zrpc_out=%s --client=%t --home %s -m --style %s",
				v,
				includeArgs,
				pbOutDir,
				pbOutDir,
				".",
				config.C.Gen.RpcClient,
				goctlHome,
				config.C.Style)

			for _, exp := range excludeThirdPartyProtoFiles {
				_, expRel, err := relToProtoDir(exp, protoDirs)
				if err != nil {
					return nil, err
				}

				expFds, err := protoParser.ParseFiles(expRel)
				if err != nil {
					return nil, err
				}
				if len(expFds) == 0 {
					continue
				}

				// M 选项的 key 需与 proto import 路径一致（buf 布局多为 user/v1/xxx.proto）
				importName, err := protoImportName(exp, protoDirs)
				if err != nil {
					importName = filepath.ToSlash(expRel)
				}

				goPackage := expFds[0].AsFileDescriptorProto().GetOptions().GetGoPackage()
				mapped := resolveGoPackageImport(jr.Module, goPackage)
				if isExternalGoPackage(goPackage) {
					command += fmt.Sprintf(" --go_opt=M%s=%s --go-grpc_opt=M%s=%s",
						importName, mapped, importName, mapped)
				} else {
					command += fmt.Sprintf(" --go_opt=module=%s --go_opt=M%s=%s --go-grpc_opt=module=%s --go-grpc_opt=M%s=%s",
						jr.Module, importName, mapped, jr.Module, importName, mapped)
				}
			}

			_ = protoRoot // used via include args / rel resolution

			_, err = execx.Run(command, config.C.Wd())
			if err != nil {
				return nil, err
			}

			// goctl zrpc_out=. 会在已有 cmd/ 脚手架项目里再吐一份入口；清掉以免污染
			jr.cleanupGoctlFrameArtifacts(v)

			if progressChan != nil {
				progressChan <- progress.NewFile(v)
				if config.C.Debug {
					progressChan <- progress.NewDebug(command)
				}
			}
		}

		for _, file := range allServerFiles {
			if filepath.Clean(file.DescFilepath) == filepath.Clean(v) {
				if _, ok := genCodeProtoSpecMap[file.DescFilepath]; ok {
					if err = jr.removeServerSuffix(file.Path); err != nil {
						goctlconsole.Warning("[warning]: remove server suffix %s meet error %v", file.Path, err)
						continue
					}
				}
			}
		}

		for _, file := range allLogicFiles {
			if _, ok := genCodeProtoSpecMap[file.DescFilepath]; ok {
				if err := jr.removeLogicSuffix(file.Path); err != nil {
					goctlconsole.Warning("[warning]: remove logic suffix %s meet error %v", file.Path, err)
					continue
				}
			}
		}

		if lo.Contains(genCodeProtoFiles, v) {
			for _, file := range allLogicFiles {
				if err = jr.changeLogicTypes(file); err != nil {
					goctlconsole.Warning("[warning]: change logic types %s meet error %v", file.Path, err)
					continue
				}
			}
		}

		// descriptor set：仅本地 pb（相对 go_package）；外部 contracts 的 descriptor 归外部管线
		if lo.Contains(genCodeProtoFiles, v) && !fileExternal {
			if pzerodesc.IsNeedGenProtoDescriptor(protoSpecMap[v]) {
				if !pathx.FileExists(pzerodesc.GetProtoDescriptorPath(v)) {
					_ = os.MkdirAll(filepath.Dir(pzerodesc.GetProtoDescriptorPath(v)), 0o755)
				}
				protocCommand := fmt.Sprintf("protoc --include_imports%s --descriptor_set_out=%s %s",
					buildProtocIncludeArgs(protoDirs),
					pzerodesc.GetProtoDescriptorPath(v),
					v,
				)
				_, err = execx.Run(protocCommand, config.C.Wd())
				if err != nil {
					return nil, err
				}
			}
		}

		for _, s := range protoSpecMap[v].Service {
			serverImports = append(serverImports, fmt.Sprintf(`%ssvr "%s/internal/server/%s"`, strings.ToLower(s.Name), jr.Module, strings.ToLower(s.Name)))
			pbPkgName := filepath.Base(strings.TrimPrefix(protoSpecMap[v].GoPackage, "./"))
			registerServers = append(registerServers, fmt.Sprintf("%s.Register%sServer(grpcServer, %ssvr.New%s(ctx))", pbPkgName, stringx.FirstUpper(s.Name), strings.ToLower(s.Name), stringx.FirstUpper(stringx.ToCamel(s.Name))))
		}
		pbImports = append(pbImports, fmt.Sprintf(`"%s"`, resolveGoPackageImport(jr.Module, protoSpecMap[v].GoPackage)))
	}

	// 同 go_package 多 proto/service 时去重 import
	serverImports = lo.Uniq(serverImports)
	pbImports = lo.Uniq(pbImports)
	registerServers = lo.Uniq(registerServers)

	if len(protoFiles) > 0 {
		if err = jr.genServer(serverImports, pbImports, registerServers); err != nil {
			return nil, err
		}
		// 无 service 的公共 proto：只为相对 go_package 生成本地 pb
		for _, dir := range protoDirs {
			if err = jr.genNoRpcServiceExcludeThirdPartyProto(dir); err != nil {
				return nil, err
			}
		}
		if err = jr.genApiMiddlewares(protoFiles); err != nil {
			return nil, err
		}
	}

	return protoSpecMap, nil
}

func findRpcServiceProtoFilesInDirs(dirs []string) ([]string, error) {
	var all []string
	for _, dir := range dirs {
		files, err := pzerodesc.FindRpcServiceProtoFiles(dir)
		if err != nil {
			return nil, err
		}
		all = append(all, files...)
	}
	return lo.Uniq(all), nil
}

func findExcludeThirdPartyProtoFilesInDirs(dirs []string) ([]string, error) {
	var all []string
	for _, dir := range dirs {
		files, err := pzerodesc.FindExcludeThirdPartyProtoFiles(dir)
		if err != nil {
			return nil, err
		}
		all = append(all, files...)
	}
	return lo.Uniq(all), nil
}

func buildProtoImportPaths(protoDirs []string) []string {
	var paths []string
	seen := map[string]struct{}{}
	add := func(p string) {
		p = filepath.Clean(p)
		if p == "" || p == "." {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}

	for _, dir := range protoDirs {
		add(dir)
		add(filepath.Join(dir, "third_party"))
		// buf module 常见布局：contracts/proto/<domain>/v1/*.proto，import 根在 contracts/proto
		add(filepath.Dir(dir))
	}
	for _, inc := range config.C.Gen.ProtoInclude {
		add(inc)
	}
	return paths
}

func buildProtocIncludeArgs(protoDirs []string) string {
	var b strings.Builder
	for _, p := range buildProtoImportPaths(protoDirs) {
		b.WriteString(" -I")
		b.WriteString(p)
	}
	return b.String()
}

func relToProtoDir(file string, protoDirs []string) (protoRoot, rel string, err error) {
	file = filepath.Clean(file)
	var best string
	for _, root := range buildProtoImportPaths(protoDirs) {
		candidate, relErr := filepath.Rel(root, file)
		if relErr != nil {
			continue
		}
		if candidate == ".." || strings.HasPrefix(candidate, ".."+string(os.PathSeparator)) {
			continue
		}

		candidate = filepath.ToSlash(candidate)
		if best == "" || len(candidate) > len(best) {
			protoRoot = root
			best = candidate
		}
	}
	if best != "" {
		return protoRoot, best, nil
	}
	return "", "", fmt.Errorf("proto file %s is not under configured proto-dir or include paths %v", file, protoDirs)
}

// protoImportName 返回用于 protoc M 选项 / import 的路径：在所有 -I 根中取最长 rel（通常对应 buf module 根）。
func protoImportName(file string, protoDirs []string) (string, error) {
	_, name, err := relToProtoDir(file, protoDirs)
	return name, err
}

func resolveGoPackageImport(module, goPackage string) string {
	goPackage = strings.TrimSpace(goPackage)
	if goPackage == "" {
		return filepath.ToSlash(filepath.Join(module, "internal", "types"))
	}
	if isExternalGoPackage(goPackage) {
		return goPackage
	}
	if strings.HasPrefix(goPackage, module) {
		return goPackage
	}
	return filepath.ToSlash(filepath.Join(module, "internal", strings.TrimPrefix(goPackage, "./")))
}

// isExternalGoPackage：绝对 go_package（非 ./ 相对路径）表示 PB 由外部管线生成，pzero 只出 stub。
func isExternalGoPackage(goPackage string) bool {
	goPackage = strings.TrimSpace(goPackage)
	return goPackage != "" && !strings.HasPrefix(goPackage, ".")
}

// cleanupGoctlFrameArtifacts 删除 goctl 在已有 pzero 项目中多余生成的入口/配置。
func (jr *PzeroRpc) cleanupGoctlFrameArtifacts(protoFile string) {
	base := strings.TrimSuffix(filepath.Base(protoFile), filepath.Ext(protoFile))
	wd := config.C.Wd()
	candidates := []string{
		filepath.Join(wd, base+".go"),
		filepath.Join(wd, "etc", base+".yaml"),
	}
	// 已有 cmd/ 脚手架时，根目录 main 由 goctl 生成的同名文件应删除
	if pathx.FileExists(filepath.Join(wd, "cmd")) {
		for _, p := range candidates {
			if pathx.FileExists(p) {
				_ = os.Remove(p)
			}
		}
	}
	// go_zero style 会再写 service_context.go，与模板 servicecontext.go 冲突
	goZeroSvc := filepath.Join(wd, "internal", "svc", "service_context.go")
	legacySvc := filepath.Join(wd, "internal", "svc", "servicecontext.go")
	if pathx.FileExists(goZeroSvc) && pathx.FileExists(legacySvc) {
		_ = os.Remove(goZeroSvc)
	}
}
