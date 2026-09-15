package genapi

import (
	"bytes"
	"go/ast"
	goparser "go/parser"
	"go/printer"
	"go/token"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	zeroconfig "github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util"

	"github.com/polpo-space/pzero/cmd/pzero/internal/config"
	jgogen "github.com/polpo-space/pzero/cmd/pzero/internal/pkg/gogen"
	"github.com/polpo-space/pzero/cmd/pzero/internal/pkg/mod"
)

func (ja *PzeroApi) getRoutesGoBody(fp string, apiSpecMap map[string]*spec.ApiSpec, currentRoutesMap map[string][]spec.Route) (string, error) {
	rootPkg, err := mod.GetParentPackage(config.C.Wd())
	if err != nil {
		return "", err
	}
	project, err := mod.GetGoMod(config.C.Wd())
	if err != nil {
		return "", err
	}

	// 获取当前文件的路由（不包含 import）
	currentRoutes := currentRoutesMap[fp]
	if len(currentRoutes) == 0 {
		return "", nil
	}

	// 创建过滤后的 ApiSpec，只包含当前文件的路由
	apiSpec := apiSpecMap[fp]
	filteredSpec := &spec.ApiSpec{
		Info:    apiSpec.Info,
		Syntax:  apiSpec.Syntax,
		Imports: apiSpec.Imports,
		Types:   apiSpec.Types, // 保留所有 types（包括 import 的）
		Service: spec.Service{
			Name:   apiSpec.Service.Name,
			Groups: []spec.Group{},
		},
	}

	// 按路由路径和方法创建映射，用于快速查找
	routeKeyMap := make(map[string]bool)
	for _, r := range currentRoutes {
		key := r.Path + ":" + r.Method
		routeKeyMap[key] = true
	}

	// 过滤路由，只保留当前文件的
	for _, g := range apiSpec.Service.Groups {
		filteredGroup := spec.Group{
			Annotation: g.Annotation,
			Routes:     []spec.Route{},
		}
		for _, r := range g.Routes {
			key := r.Path + ":" + r.Method
			if routeKeyMap[key] {
				filteredGroup.Routes = append(filteredGroup.Routes, r)
			}
		}
		if len(filteredGroup.Routes) > 0 {
			filteredSpec.Service.Groups = append(filteredSpec.Service.Groups, filteredGroup)
		}
	}

	// 使用过滤后的 ApiSpec 生成路由代码
	routesGoBody, err := jgogen.GenRoutesString(rootPkg, project.Path, &zeroconfig.Config{NamingFormat: config.C.Style}, filteredSpec)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()
	f, err := goparser.ParseFile(fset, "", strings.NewReader(routesGoBody), goparser.ParseComments)
	if err != nil {
		return "", err
	}

	// 只处理当前文件的路由
	for _, route := range currentRoutes {
		ast.Inspect(f, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
					if _, ok := sel.X.(*ast.Ident); ok {
						if util.Title(sel.Sel.Name) == util.Title(strings.TrimSuffix(route.Handler, "Handler"))+"Handler" {
							sel.Sel.Name = strings.TrimSuffix(route.Handler, "Handler")
							// 查找对应的 group
							for _, g := range filteredSpec.Service.Groups {
								for _, gr := range g.Routes {
									if gr.Handler == route.Handler && g.GetAnnotation("group") != "" {
										sel.Sel.Name = util.Title(strings.TrimSuffix(route.Handler, "Handler"))
										break
									}
								}
							}
						}
					}
				} else if indent, ok := n.Fun.(*ast.Ident); ok {
					if util.Title(indent.Name) == util.Title(strings.TrimSuffix(route.Handler, "Handler"))+"Handler" {
						indent.Name = strings.TrimSuffix(route.Handler, "Handler")
						for _, g := range filteredSpec.Service.Groups {
							for _, gr := range g.Routes {
								if gr.Handler == route.Handler && g.GetAnnotation("group") != "" {
									indent.Name = util.Title(strings.TrimSuffix(route.Handler, "Handler"))
									break
								}
							}
						}
					}
				}
			}
			return true
		})
	}

	// 遍历 AST 节点
	for _, decl := range f.Decls {
		// 查找函数声明
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "RegisterHandlers" {
				// 提取函数体
				var buf bytes.Buffer
				if err = printer.Fprint(&buf, fset, funcDecl.Body); err != nil {
					return "", err
				}
				return buf.String(), nil
			}
		}
	}

	return "", nil
}
