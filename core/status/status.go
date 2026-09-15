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

// errorInfoDomain 标识 ErrorInfo detail 由 pzero 业务错误码写入。
const errorInfoDomain = "pzero"

// grpcCodes 业务码到 gRPC code 的映射, 未列出的业务码一律为 Unknown。
// 只有映射到 Internal 的错误会触发 go-zero 熔断统计, 普通业务错误不会。
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

// httpCodes gRPC code 到业务码的反向映射, 用于没有携带 ErrorInfo 的普通 gRPC 错误。
var httpCodes = func() map[codes.Code]Code {
	m := make(map[codes.Code]Code, len(grpcCodes))
	for h, g := range grpcCodes {
		m[g] = h
	}
	return m
}()

type Status struct {
	code    Code
	message string
	err     error
	extra   any
}

var (
	statusMap = map[Code]Status{}
	mu        sync.RWMutex
)

func Register(code Code) {
	RegisterWithMessage(code, "")
}

func RegisterWithMessage(code Code, message string) {
	mu.Lock()
	defer mu.Unlock()
	statusMap[code] = Status{
		code:    code,
		message: message,
	}
}

func New(code Code, message string, err error) *Status {
	return &Status{
		code:    code,
		message: message,
		err:     err,
	}
}

func Error(code Code) error {
	mu.RLock()
	defer mu.RUnlock()

	status, ok := statusMap[code]
	if ok {
		return errors.WithStack(status)
	}
	return Error(http.StatusInternalServerError)
}

func ErrorMessage(code Code, message string) error {
	mu.RLock()
	defer mu.RUnlock()

	status, ok := statusMap[code]
	if ok {
		status.message = message
		return errors.WithStack(status)
	}
	return Error(http.StatusInternalServerError)
}

func Wrap(code Code, err error, extra ...any) error {
	mu.RLock()
	defer mu.RUnlock()

	status, ok := statusMap[code]
	if ok {
		status.err = err
		if len(extra) == 1 {
			status.extra = extra[0]
		}
		return errors.WithStack(status)
	}
	return Error(http.StatusInternalServerError)
}

// FromError 从任意 error 还原 Status。
// 识别顺序: 本地 Status -> gRPC status(优先读取 ErrorInfo 中的业务码, 否则按 gRPC code 反向映射) -> 500。
func FromError(err error) *Status {
	err = errors.Cause(err)
	var status Status
	if errors.As(err, &status) {
		return &status
	}
	if st, ok := grpcstatus.FromError(err); ok && st != nil {
		return fromGRPC(st, err)
	}
	return New(http.StatusInternalServerError, "", err)
}

func fromGRPC(st *grpcstatus.Status, err error) *Status {
	for _, detail := range st.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if !ok || info.GetDomain() != errorInfoDomain {
			continue
		}
		if code, parseErr := strconv.Atoi(info.GetReason()); parseErr == nil {
			return New(Code(code), st.Message(), nil)
		}
	}
	code, ok := httpCodes[st.Code()]
	if !ok {
		code = http.StatusInternalServerError
	}
	return New(code, st.Message(), err)
}

// GRPCStatus 使 Status 可直接作为 gRPC handler 的返回错误:
// gRPC code 取自映射表, message 与 Error() 一致, 业务码写入 ErrorInfo.Reason 随 details 传输。
// grpc-go 的 status.FromError 会沿 Unwrap 链识别该方法, 因此经 errors.WithStack 包装后同样生效。
func (e Status) GRPCStatus() *grpcstatus.Status {
	code, ok := grpcCodes[e.code]
	if !ok {
		code = codes.Unknown
	}
	st := grpcstatus.New(code, e.Error())
	withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
		Reason: strconv.Itoa(int(e.code)),
		Domain: errorInfoDomain,
	})
	if err != nil {
		return st
	}
	return withDetails
}

func (e Status) Error() string {
	message := e.message
	if e.err != nil {
		if message == "" {
			return e.err.Error()
		}
		message = message + ": " + e.err.Error()
	}
	return message
}

func (e Status) Unwrap() error {
	return e.err
}

func (e Status) Extra() any {
	return e.extra
}

func (e Status) Code() Code {
	return e.code
}

func (e Status) Message() string {
	return e.message
}

func init() {
	Register(http.StatusInternalServerError)
}
