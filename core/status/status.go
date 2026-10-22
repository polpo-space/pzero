package status

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type Code int

const (
	// errorInfoDomain 标识 ErrorInfo detail 由 pzero 业务错误码写入。
	errorInfoDomain = "pzero"
	// defaultMessage 任何 code 在没有显式提示、注册提示与 HTTP 标准文案时的最终回退, 保证对外永不为空也永不泄露 cause。
	defaultMessage = "internal server error"
)

// grpcCodes 具备 HTTP 语义的业务码到 gRPC code 的映射; 未列出且未通过 WithGRPCCode 指定的业务码一律为 Unknown。
// go-zero zrpc 熔断只把 DeadlineExceeded/Internal/Unavailable/DataLoss/Unimplemented/ResourceExhausted 计为失败,
// 因此 429/500/501/503/504 会计入熔断, 其余业务错误不会。
var grpcCodes = map[Code]codes.Code{
	http.StatusBadRequest:          codes.InvalidArgument,
	http.StatusUnauthorized:        codes.Unauthenticated,
	http.StatusForbidden:           codes.PermissionDenied,
	http.StatusNotFound:            codes.NotFound,
	http.StatusConflict:            codes.AlreadyExists,
	http.StatusTooManyRequests:     codes.ResourceExhausted,
	http.StatusInternalServerError: codes.Internal,
	http.StatusNotImplemented:      codes.Unimplemented,
	http.StatusServiceUnavailable:  codes.Unavailable,
	http.StatusGatewayTimeout:      codes.DeadlineExceeded,
}

// httpCodes gRPC code 到业务码的反向映射, 用于没有携带 ErrorInfo 的普通 gRPC 错误(如旧服务手写的 grpc/status)。
// 与 grpc-gateway 的 HTTPStatusFromCode 保持一致; 未列出的 code 回退为 500。
var httpCodes = map[codes.Code]Code{
	codes.Canceled:           499,
	codes.Unknown:            http.StatusInternalServerError,
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.DeadlineExceeded:   http.StatusGatewayTimeout,
	codes.NotFound:           http.StatusNotFound,
	codes.AlreadyExists:      http.StatusConflict,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.Unauthenticated:    http.StatusUnauthorized,
	codes.ResourceExhausted:  http.StatusTooManyRequests,
	codes.FailedPrecondition: http.StatusBadRequest,
	codes.Aborted:            http.StatusConflict,
	codes.OutOfRange:         http.StatusBadRequest,
	codes.Unimplemented:      http.StatusNotImplemented,
	codes.Internal:           http.StatusInternalServerError,
	codes.Unavailable:        http.StatusServiceUnavailable,
	codes.DataLoss:           http.StatusInternalServerError,
}

// Status 业务错误。
// Error() 面向日志与 Go error chain, 可包含底层 cause;
// Message() 面向客户端, 永不包含 cause; GRPCStatus() 与 HTTP 响应只应使用 Message()。
type Status struct {
	code     Code
	message  string
	err      error
	extra    any
	grpcCode codes.Code // 注册时显式指定的 gRPC code, OK 表示未指定
	origin   Origin
}

// Origin 表示 FromError 还原出的 Status 从哪里来, 调用方据此决定 code 与 Message() 是否可信。
type Origin int

const (
	// OriginLocal 本进程通过 Error / ErrorMessage / Wrap / New 创建。零值, 注册与创建无需设置。
	OriginLocal Origin = iota
	// OriginRemote 从携带 pzero ErrorInfo 的 gRPC 错误还原: code 是下游注册的业务码,
	// Message() 是下游服务写的文案, 是否展示给用户由调用方决定。
	OriginRemote
	// OriginGRPC 从不带 ErrorInfo 的普通 gRPC 错误还原: code 由 gRPC code 反向映射推断, Message() 未经审核。
	OriginGRPC
	// OriginUnknown 既不是 Status 也不是 gRPC 错误, 统一为 500。
	OriginUnknown
)

// Option 注册错误码时的可选配置。
type Option func(*Status)

// WithMessage 指定该错误码的默认对外提示。
func WithMessage(message string) Option {
	return func(s *Status) {
		s.message = message
	}
}

// WithGRPCCode 为业务码显式指定 gRPC code, 覆盖默认映射(未映射的业务码默认为 Unknown)。
func WithGRPCCode(code codes.Code) Option {
	return func(s *Status) {
		s.grpcCode = code
	}
}

var (
	statusMap = map[Code]Status{}
	mu        sync.RWMutex
)

func Register(code Code, opts ...Option) {
	s := Status{code: code}
	for _, opt := range opts {
		opt(&s)
	}
	mu.Lock()
	defer mu.Unlock()
	statusMap[code] = s
}

func RegisterWithMessage(code Code, message string) {
	Register(code, WithMessage(message))
}

// Lookup 返回 code 注册时的 Status; 未注册时 ok 为 false。
// 返回值是副本, Message() 为注册的默认提示, 用于按码查文案或在测试中校验注册是否齐全。
func Lookup(code Code) (*Status, bool) {
	status, ok := lookup(code)
	if !ok {
		return nil, false
	}
	return &status, true
}

func New(code Code, message string, err error) *Status {
	return &Status{
		code:    code,
		message: message,
		err:     err,
	}
}

func lookup(code Code) (Status, bool) {
	mu.RLock()
	defer mu.RUnlock()
	status, ok := statusMap[code]
	return status, ok
}

func Error(code Code) error {
	status, ok := lookup(code)
	if ok {
		// 不能用 errors.WithStack 包装: grpc-go 对 wrap 过的 status 会把 p.Message 改成 err.Error(),
		// 从而把 Error() 里的 cause 泄漏到 gRPC message。Status 自身必须直接实现 GRPCStatus。
		return status
	}
	return Error(http.StatusInternalServerError)
}

func ErrorMessage(code Code, message string) error {
	status, ok := lookup(code)
	if ok {
		status.message = message
		return status
	}
	return Error(http.StatusInternalServerError)
}

func Wrap(code Code, err error, extra ...any) error {
	status, ok := lookup(code)
	if ok {
		status.err = err
		if len(extra) == 1 {
			status.extra = extra[0]
		}
		return status
	}
	return Error(http.StatusInternalServerError)
}

// FromError 从任意 error 还原 Status, 来源见 Origin()。
// 识别顺序: 本地 Status(值或 New 返回的指针) -> gRPC status(优先读取 ErrorInfo 中的业务码, 否则按 gRPC code 反向映射) -> 500。
func FromError(err error) *Status {
	err = errors.Cause(err)
	var status Status
	if errors.As(err, &status) {
		return &status
	}
	var pointer *Status
	if errors.As(err, &pointer) && pointer != nil {
		status = *pointer
		return &status
	}
	if st, ok := grpcstatus.FromError(err); ok && st != nil {
		return fromGRPC(st, err)
	}
	unknown := New(http.StatusInternalServerError, "", err)
	unknown.origin = OriginUnknown
	return unknown
}

func fromGRPC(st *grpcstatus.Status, err error) *Status {
	for _, detail := range st.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if !ok || info.GetDomain() != errorInfoDomain {
			continue
		}
		if code, parseErr := strconv.Atoi(info.GetReason()); parseErr == nil {
			remote := New(Code(code), st.Message(), nil)
			remote.origin = OriginRemote
			return remote
		}
	}
	code, ok := httpCodes[st.Code()]
	if !ok {
		code = http.StatusInternalServerError
	}
	mapped := New(code, st.Message(), err)
	mapped.origin = OriginGRPC
	return mapped
}

// GRPCStatus 使 Status 可直接作为 gRPC handler 的返回错误:
// gRPC code 优先取注册时指定值, 其次映射表, 否则 Unknown; message 只传 Message(), 底层 cause 不会离开本进程;
// 业务码写入 ErrorInfo.Reason 随 details 传输。
// grpc-go 的 status.FromError 会沿 Unwrap 链识别该方法, 因此经 errors.WithStack 包装后同样生效。
func (e Status) GRPCStatus() *grpcstatus.Status {
	code := e.grpcCode
	if code == codes.OK {
		var ok bool
		if code, ok = grpcCodes[e.code]; !ok {
			code = codes.Unknown
		}
	}
	st := grpcstatus.New(code, e.Message())
	withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
		Reason: strconv.Itoa(int(e.code)),
		Domain: errorInfoDomain,
	})
	if err != nil {
		return st
	}
	return withDetails
}

// Error 面向日志与 error chain 的完整描述, 包含底层 cause; 不要用它构造对外响应。
func (e Status) Error() string {
	message := e.Message()
	if e.err != nil {
		return message + ": " + e.err.Error()
	}
	return message
}

func (e Status) Unwrap() error {
	return e.err
}

func (e Status) Extra() any {
	return e.extra
}

// Origin 返回该 Status 的来源, 见 Origin 各取值说明。
func (e Status) Origin() Origin {
	return e.origin
}

func (e Status) Code() Code {
	return e.code
}

// Message 面向客户端的安全提示, 永不包含底层 cause。
// 顺序: 显式 message -> 该 code 注册时的默认提示 -> HTTP 标准文案 -> "internal server error"。
func (e Status) Message() string {
	if e.message != "" {
		return e.message
	}
	if registered, ok := lookup(e.code); ok && registered.message != "" {
		return registered.message
	}
	if text := http.StatusText(int(e.code)); text != "" {
		return text
	}
	return defaultMessage
}

// init 预注册 grpcCodes 中有 HTTP 语义的通用码, 应用无需为使用 400/404 等再注册一遍。
// 应用可以用 Register 覆盖默认提示(导入包的 init 先于应用的变量初始化执行);
// 以后若对重复注册做校验, 必须继续允许覆盖这些内置默认值。
func init() {
	for code := range grpcCodes {
		RegisterWithMessage(code, http.StatusText(int(code)))
	}
	RegisterWithMessage(http.StatusInternalServerError, defaultMessage)
}
