package genapi

import (
	"errors"
	goparser "go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

func TestPatchLogicKeepsExistingLogicWhenRewriteHandlerFalse(t *testing.T) {
	tmpDir := withTempWorkDir(t)

	logicDir := filepath.Join(tmpDir, "internal", "logic", "user")
	if err := os.MkdirAll(logicDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	existingPath := filepath.Join(logicDir, "login.go")
	existingContent := []byte(`package user

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"example.com/app/internal/svc"
)

type Login struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
	w      http.ResponseWriter
}

func NewLogin(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request, w http.ResponseWriter) *Login {
	return &Login{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
		w:      w,
	}
}

func (l *Login) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	return nil, nil
}
`)
	if err := os.WriteFile(existingPath, existingContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	generatedPath := filepath.Join(logicDir, "login_logic.go")
	if err := os.WriteFile(generatedPath, []byte("package user\n\ntype LoginLogic struct{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	apiFile := filepath.Join("desc", "api", "user.api")
	ja := &PzeroApi{}
	err := ja.patchLogic(LogicFile{
		Path:           generatedPath,
		DescFilepath:   apiFile,
		Handler:        "Login",
		RewriteHandler: false,
		RequestType:    spec.DefineStruct{RawName: "LoginRequest"},
		ResponseType:   spec.DefineStruct{RawName: "LoginResponse"},
	}, map[string]*spec.ApiSpec{apiFile: {}})
	if err != nil {
		t.Fatalf("patchLogic() error = %v", err)
	}

	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != string(existingContent) {
		t.Fatalf("patchLogic() should keep existing logic unchanged, got:\n%s", data)
	}
	if _, err = os.Stat(generatedPath); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("patchLogic() should remove generated logic when final logic exists, stat err = %v", err)
	}
}

func TestPatchLogicPreservesSSESignature(t *testing.T) {
	tmpDir := withTempWorkDir(t)
	setTypesDir(t, defaultTypesDir)

	logicDir := filepath.Join(tmpDir, "internal", "logic", "conversation")
	if err := os.MkdirAll(logicDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	generatedPath := filepath.Join(logicDir, "streamevents_logic.go")
	generatedContent := []byte(`package conversation

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"example.com/app/internal/svc"
	"example.com/app/internal/types"
)

type StreamEventsLogic struct {
	logx.Logger
	ctx context.Context
	svcCtx *svc.ServiceContext
}

func NewStreamEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StreamEventsLogic {
	return &StreamEventsLogic{
		Logger: logx.WithContext(ctx),
		ctx: ctx,
		svcCtx: svcCtx,
	}
}

func (l *StreamEventsLogic) StreamEvents(req *types.EventsRequest, client chan<- *types.EventsResponse) error {
	return nil
}
`)
	if err := os.WriteFile(generatedPath, generatedContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	apiFile := filepath.Join("desc", "api", "conversation.api")
	ja := &PzeroApi{Module: "example.com/app"}
	err := ja.patchLogic(LogicFile{
		Package:        "conversation",
		Path:           generatedPath,
		DescFilepath:   apiFile,
		Handler:        "StreamEvents",
		RewriteHandler: true,
		SSE:            true,
		RequestType:    spec.DefineStruct{RawName: "EventsRequest"},
		ResponseType:   spec.DefineStruct{RawName: "EventsResponse"},
	}, map[string]*spec.ApiSpec{apiFile: {}})
	if err != nil {
		t.Fatalf("patchLogic() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(logicDir, "streamevents.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	for _, want := range []string{
		"type StreamEvents struct",
		"func NewStreamEvents(",
		"client chan<- *types.EventsResponse",
		`types "example.com/app/internal/types/conversation"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("patchLogic() missing %q in SSE logic:\n%s", want, got)
		}
	}
	if strings.Contains(got, "net/http") {
		t.Fatalf("patchLogic() should not add HTTP dependencies to SSE logic:\n%s", got)
	}
}

func TestPatchLogicPreservesNonLogicReceiverName(t *testing.T) {
	tmpDir := withTempWorkDir(t)
	setTypesDir(t, defaultTypesDir)

	logicDir := filepath.Join(tmpDir, "internal", "logic", "asset")
	if err := os.MkdirAll(logicDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	generatedPath := filepath.Join(logicDir, "importimage_logic.go")
	generatedContent := []byte(`package asset

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"example.com/app/internal/svc"
	"example.com/app/internal/types"
)

type httpMappedError struct {
	err error
}

func (e *httpMappedError) Error() string {
	return e.err.Error()
}

type ImportImageLogic struct {
	logx.Logger
	ctx context.Context
	svcCtx *svc.ServiceContext
}

func NewImportImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ImportImageLogic {
	return &ImportImageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ImportImageLogic) ImportImage(req *types.ImportImageRequest) (resp *types.ImportImageResponse, err error) {
	return nil, errors.New("not implemented")
}
`)
	if err := os.WriteFile(generatedPath, generatedContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	apiFile := filepath.Join("desc", "api", "asset.api")
	ja := &PzeroApi{Module: "example.com/app"}
	if err := ja.patchLogic(LogicFile{
		Package:        "asset",
		Path:           generatedPath,
		DescFilepath:   apiFile,
		Handler:        "ImportImage",
		RewriteHandler: true,
		RequestType:    spec.DefineStruct{RawName: "ImportImageRequest"},
		ResponseType:   spec.DefineStruct{RawName: "ImportImageResponse"},
	}, map[string]*spec.ApiSpec{apiFile: {}}); err != nil {
		t.Fatalf("patchLogic() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(logicDir, "importimage.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "func (e *httpMappedError) Error() string") {
		t.Fatalf("patchLogic() changed a non-Logic receiver name:\n%s", got)
	}
}

func TestPatchLogicPreservesUnnamedResultParameters(t *testing.T) {
	tmpDir := withTempWorkDir(t)
	setTypesDir(t, defaultTypesDir)

	logicDir := filepath.Join(tmpDir, "internal", "logic", "user")
	if err := os.MkdirAll(logicDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	existingPath := filepath.Join(logicDir, "get.go")
	existingContent := []byte(`package user

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"example.com/app/internal/svc"
	"example.com/app/internal/types"
)

type Get struct {
	logx.Logger
	ctx context.Context
	svcCtx *svc.ServiceContext
	r *http.Request
}

func NewGet(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Get {
	return &Get{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx, r: r}
}

func (l *Get) Get(req *types.GetRequest) (*types.GetResponse, error) {
	resp, err := loadUser(req)
	return resp, err
}
`)
	if err := os.WriteFile(existingPath, existingContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	generatedPath := filepath.Join(logicDir, "get_logic.go")
	if err := os.WriteFile(generatedPath, []byte("package user\n\ntype GetLogic struct{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	apiFile := filepath.Join("desc", "api", "user.api")
	ja := &PzeroApi{Module: "example.com/app"}
	if err := ja.patchLogic(LogicFile{
		Package:        "user",
		Path:           generatedPath,
		DescFilepath:   apiFile,
		Handler:        "Get",
		RewriteHandler: true,
		RequestType:    spec.DefineStruct{RawName: "GetRequest"},
		ResponseType:   spec.DefineStruct{RawName: "GetResponse"},
	}, map[string]*spec.ApiSpec{apiFile: {}}); err != nil {
		t.Fatalf("patchLogic() error = %v", err)
	}

	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "Get(req *types.GetRequest) (*types.GetResponse, error)") {
		t.Fatalf("patchLogic() changed unnamed result parameters:\n%s", got)
	}
}

func TestPatchLogicUpdatesExistingLogicInDifferentlyNamedFile(t *testing.T) {
	tmpDir := withTempWorkDir(t)
	setTypesDir(t, defaultTypesDir)

	logicDir := filepath.Join(tmpDir, "internal", "logic", "order")
	if err := os.MkdirAll(logicDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	existingPath := filepath.Join(logicDir, "order_shipping.go")
	existingContent := []byte(`package order

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"example.com/app/internal/svc"
	"example.com/app/internal/types"
)

type OrderMarkShipped struct {
	logx.Logger
	ctx context.Context
	svcCtx *svc.ServiceContext
	r *http.Request
}

func NewOrderMarkShipped(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *OrderMarkShipped {
	return &OrderMarkShipped{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx, r: r}
}

func (l *OrderMarkShipped) OrderMarkShipped(req *types.OldRequest) (*types.OldResponse, error) {
	return nil, nil
}
`)
	if err := os.WriteFile(existingPath, existingContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	generatedPath := filepath.Join(logicDir, "order_mark_shipped_logic.go")
	if err := os.WriteFile(generatedPath, []byte("package order\n\ntype OrderMarkShippedLogic struct{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	apiFile := filepath.Join("desc", "api", "order.api")
	ja := &PzeroApi{Module: "example.com/app"}
	if err := ja.patchLogic(LogicFile{
		Package:        "order",
		Path:           generatedPath,
		DescFilepath:   apiFile,
		Handler:        "OrderMarkShipped",
		RewriteHandler: true,
		RequestType:    spec.DefineStruct{RawName: "OrderMarkShippedRequest"},
		ResponseType:   spec.DefineStruct{RawName: "OrderMarkShippedResponse"},
	}, map[string]*spec.ApiSpec{apiFile: {}}); err != nil {
		t.Fatalf("patchLogic() error = %v", err)
	}

	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "OrderMarkShipped(req *types.OrderMarkShippedRequest) (*types.OrderMarkShippedResponse, error)") {
		t.Fatalf("patchLogic() did not update the existing grouped logic:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(logicDir, "order_mark_shipped.go")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("patchLogic() should not create a duplicate method-level logic file, stat err = %v", err)
	}
}

func TestPatchLogicKeepsDifferentlyNamedLogicWhenRewriteHandlerFalse(t *testing.T) {
	tmpDir := withTempWorkDir(t)

	logicDir := filepath.Join(tmpDir, "internal", "logic", "order")
	if err := os.MkdirAll(logicDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	existingPath := filepath.Join(logicDir, "order_shipping.go")
	existingContent := []byte(`package order

type OrderMarkShipped struct{}

func (l *OrderMarkShipped) OrderMarkShipped() error {return nil}
`)
	if err := os.WriteFile(existingPath, existingContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	generatedPath := filepath.Join(logicDir, "order_mark_shipped_logic.go")
	if err := os.WriteFile(generatedPath, []byte("package order\n\ntype OrderMarkShippedLogic struct{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	ja := &PzeroApi{}
	if err := ja.patchLogic(LogicFile{
		Path:           generatedPath,
		Handler:        "OrderMarkShipped",
		RewriteHandler: false,
	}, nil); err != nil {
		t.Fatalf("patchLogic() error = %v", err)
	}

	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != string(existingContent) {
		t.Fatalf("patchLogic() should keep differently named custom logic byte-for-byte unchanged, got:\n%s", data)
	}
	if _, err := os.Stat(generatedPath); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("patchLogic() should remove generated logic, stat err = %v", err)
	}
}

func TestPatchLogicDropsResultNamesWhenResultArityChanges(t *testing.T) {
	tmpDir := withTempWorkDir(t)
	setTypesDir(t, defaultTypesDir)

	logicDir := filepath.Join(tmpDir, "internal", "logic", "user")
	if err := os.MkdirAll(logicDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	existingPath := filepath.Join(logicDir, "get.go")
	existingContent := []byte(`package user

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"example.com/app/internal/svc"
)

type Get struct {
	logx.Logger
	ctx context.Context
	svcCtx *svc.ServiceContext
	r *http.Request
}

func NewGet(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Get {
	return &Get{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx, r: r}
}

func (l *Get) Get() (err error) {
	return nil
}
`)
	if err := os.WriteFile(existingPath, existingContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	generatedPath := filepath.Join(logicDir, "get_logic.go")
	if err := os.WriteFile(generatedPath, []byte("package user\n\ntype GetLogic struct{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	apiFile := filepath.Join("desc", "api", "user.api")
	ja := &PzeroApi{Module: "example.com/app"}
	if err := ja.patchLogic(LogicFile{
		Package:        "user",
		Path:           generatedPath,
		DescFilepath:   apiFile,
		Handler:        "Get",
		RewriteHandler: true,
		ResponseType:   spec.DefineStruct{RawName: "GetResponse"},
	}, map[string]*spec.ApiSpec{apiFile: {}}); err != nil {
		t.Fatalf("patchLogic() error = %v", err)
	}

	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "Get() (*types.GetResponse, error)") {
		t.Fatalf("patchLogic() should use unnamed results when arity changes:\n%s", got)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), existingPath, data, 0); err != nil {
		t.Fatalf("generated logic is not valid Go: %v\n%s", err, got)
	}
}
