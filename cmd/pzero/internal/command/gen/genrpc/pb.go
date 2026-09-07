package genrpc

import (
	"fmt"
	"path/filepath"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/execx"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	pzerodesc "github.com/polpo-space/pzero/cmd/pzero/internal/desc"
)

func (jr *PzeroRpc) genNoRpcServiceExcludeThirdPartyProto(protoDirPath string) error {
	excludeThirdPartyProtoFiles, err := pzerodesc.FindNoRpcServiceExcludeThirdPartyProtoFiles(protoDirPath)
	if err != nil {
		return err
	}
	if len(excludeThirdPartyProtoFiles) == 0 {
		return nil
	}

	protoDirs := config.C.ProtoDirs()
	var protoParser protoparse.Parser
	protoParser.InferImportPaths = false
	protoParser.ImportPaths = buildProtoImportPaths(protoDirs)
	protoParser.IncludeSourceCodeInfo = true

	for _, v := range excludeThirdPartyProtoFiles {
		_, rel, err := relToProtoDir(v, protoDirs)
		if err != nil {
			rel, err = filepath.Rel(protoDirPath, v)
			if err != nil {
				return err
			}
		}

		fds, err := protoParser.ParseFiles(rel)
		if err != nil {
			return err
		}
		if len(fds) == 0 {
			continue
		}

		goPackage := fds[0].AsFileDescriptorProto().GetOptions().GetGoPackage()
		if isExternalGoPackage(goPackage) {
			continue
		}
		mapped := resolveGoPackageImport(jr.Module, goPackage)

		command := fmt.Sprintf("protoc %s%s --go_out=%s --go_opt=module=%s --go_opt=M%s=%s --go-grpc_out=%s --go-grpc_opt=module=%s",
			v,
			buildProtocIncludeArgs(protoDirs),
			filepath.Join("."),
			jr.Module,
			rel,
			mapped,
			filepath.Join("."),
			jr.Module)

		_, err = execx.Run(command, config.C.Wd())
		if err != nil {
			return err
		}
	}
	return nil
}
