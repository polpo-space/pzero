package genrpc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	pzeroapi "github.com/jzero-io/desc/proto/api"
	"github.com/rinchsan/gosimports"
	"github.com/samber/lo"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"google.golang.org/protobuf/proto"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	pzerodesc "github.com/polpo-space/pzero/cmd/pzero/internal/desc"
	"github.com/polpo-space/pzero/cmd/pzero/internal/embeded"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/console"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/templatex"
)

type PzeroProtoApiMiddleware struct {
	Name   string
	Routes []string
}

func (jr *PzeroRpc) genApiMiddlewares(protoFiles []string) (err error) {
	var fds []*desc.FileDescriptor

	// parse proto（与 genrpc 共用 proto-dir / import 根，支持中央 contracts）
	var protoParser protoparse.Parser

	protoParser.InferImportPaths = false

	protoDirs := config.C.ProtoDirs()
	var files []string
	for _, protoFilename := range protoFiles {
		_, rel, err := relToProtoDir(protoFilename, protoDirs)
		if err != nil {
			return err
		}
		files = append(files, rel)
	}

	protoParser.ImportPaths = buildProtoImportPaths(protoDirs)
	protoParser.IncludeSourceCodeInfo = true
	fds, err = protoParser.ParseFiles(files...)
	if err != nil {
		return err
	}

	var httpMiddlewares []PzeroProtoApiMiddleware
	var zrpcMiddlewares []PzeroProtoApiMiddleware

	httpMapMiddlewares := orderedmap.New()
	zrpcMapMiddlewares := orderedmap.New()

	for _, fd := range fds {
		descriptorProto := fd.AsFileDescriptorProto()

		var methodUrls []string
		var fullMethods []string

		for _, service := range descriptorProto.GetService() {
			for _, method := range service.GetMethod() {
				methodUrls = append(methodUrls, pzerodesc.GetRpcMethodUrl(method))
				fullMethods = append(fullMethods, fmt.Sprintf("/%s.%s/%s", fd.GetPackage(), service.GetName(), method.GetName()))

				httpGroupExt := proto.GetExtension(service.GetOptions(), pzeroapi.E_HttpGroup)
				switch rule := httpGroupExt.(type) {
				case *pzeroapi.HttpRule:
					if rule != nil {
						split := strings.Split(strings.ReplaceAll(rule.Middleware, " ", ""), ",")
						for _, m := range split {
							if urls, ok := httpMapMiddlewares.Get(m); ok {
								urls = append(urls.([]string), methodUrls...)
								httpMapMiddlewares.Set(m, urls)
							} else {
								httpMapMiddlewares.Set(m, methodUrls)
							}
						}
					}
				}

				zrpcGroupExt := proto.GetExtension(service.GetOptions(), pzeroapi.E_ZrpcGroup)
				switch rule := zrpcGroupExt.(type) {
				case *pzeroapi.ZrpcRule:
					if rule != nil {
						split := strings.Split(strings.ReplaceAll(rule.Middleware, " ", ""), ",")
						for _, m := range split {
							if fms, ok := zrpcMapMiddlewares.Get(m); ok {
								fms = append(fms.([]string), fullMethods...)
								zrpcMapMiddlewares.Set(m, fms)
							} else {
								zrpcMapMiddlewares.Set(m, fullMethods)
							}
						}
					}
				}

				httpExt := proto.GetExtension(method.GetOptions(), pzeroapi.E_Http)
				switch rule := httpExt.(type) {
				case *pzeroapi.HttpRule:
					if rule != nil {
						split := strings.Split(strings.ReplaceAll(rule.Middleware, " ", ""), ",")
						for _, m := range split {
							if urls, ok := httpMapMiddlewares.Get(m); ok {
								urls = append(urls.([]string), pzerodesc.GetRpcMethodUrl(method))
								httpMapMiddlewares.Set(m, urls)
							} else {
								httpMapMiddlewares.Set(m, []string{pzerodesc.GetRpcMethodUrl(method)})
							}
						}
					}
				}

				zrpcExt := proto.GetExtension(method.GetOptions(), pzeroapi.E_Zrpc)
				switch rule := zrpcExt.(type) {
				case *pzeroapi.ZrpcRule:
					if rule != nil {
						split := strings.Split(strings.ReplaceAll(rule.Middleware, " ", ""), ",")
						for _, m := range split {
							if urls, ok := zrpcMapMiddlewares.Get(m); ok {
								urls = append(urls.([]string), fmt.Sprintf("/%s.%s/%s", fd.GetPackage(), service.GetName(), method.GetName()))
								zrpcMapMiddlewares.Set(m, urls)
							} else {
								zrpcMapMiddlewares.Set(m, []string{fmt.Sprintf("/%s.%s/%s", fd.GetPackage(), service.GetName(), method.GetName())})
							}
						}
					}
				}
			}
		}
	}

	// order and unique and transfer to httpMiddlewares and zrpcMiddlewares
	httpMiddlewares = processMiddlewares(httpMapMiddlewares)
	zrpcMiddlewares = processMiddlewares(zrpcMapMiddlewares)

	if len(httpMiddlewares) == 0 && len(zrpcMiddlewares) == 0 {
		return nil
	}

	for _, v := range httpMiddlewares {
		template, err := templatex.ParseTemplate(filepath.Join("rpc", "middleware_http.go.tpl"), map[string]any{
			"Name": v.Name,
		}, embeded.ReadTemplateFile(filepath.Join("rpc", "middleware_http.go.tpl")))
		if err != nil {
			return err
		}
		process, err := gosimports.Process("", template, nil)
		if err != nil {
			return err
		}
		namingFormat, _ := format.FileNamingFormat(config.C.Style, v.Name+"Middleware")
		if !pathx.FileExists(filepath.Join("internal", "middleware", namingFormat+".go")) {
			err = os.WriteFile(filepath.Join("internal", "middleware", namingFormat+".go"), process, 0o644)
			if err != nil {
				return err
			}
		}
	}

	for _, v := range zrpcMiddlewares {
		template, err := templatex.ParseTemplate(filepath.Join("rpc", "middleware_zrpc.go.tpl"), map[string]any{
			"Name": v.Name,
		}, embeded.ReadTemplateFile(filepath.Join("rpc", "middleware_zrpc.go.tpl")))
		if err != nil {
			return err
		}

		process, err := gosimports.Process("", template, nil)
		if err != nil {
			return err
		}
		namingFormat, _ := format.FileNamingFormat(config.C.Style, v.Name+"Middleware")
		if !pathx.FileExists(filepath.Join("internal", "middleware", namingFormat+".go")) {
			err = os.WriteFile(filepath.Join("internal", "middleware", namingFormat+".go"), process, 0o644)
			if err != nil {
				return err
			}
		}
	}

	template, err := templatex.ParseTemplate(filepath.Join("rpc", "middleware_gen.go.tpl"), map[string]any{
		"HttpMiddlewares": httpMiddlewares,
		"ZrpcMiddlewares": zrpcMiddlewares,
	}, embeded.ReadTemplateFile(filepath.Join("rpc", "middleware_gen.go.tpl")))
	if err != nil {
		return err
	}

	process, err := gosimports.Process("", template, nil)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("internal", "middleware", "middleware_gen.go"), process, 0o644)
	if err != nil {
		return err
	}

	if !config.C.Quiet {
		fmt.Printf("%s\n", console.Green("Gen Rpc Middleware Done"))
	}
	return nil
}

func processMiddlewares(middlewareMap *orderedmap.OrderedMap) []PzeroProtoApiMiddleware {
	var result []PzeroProtoApiMiddleware

	for _, m := range middlewareMap.Keys() {
		v, _ := middlewareMap.Get(m)
		v = lo.Uniq(v.([]string))
		result = append(result, PzeroProtoApiMiddleware{Name: m, Routes: v.([]string)})
	}
	return result
}
