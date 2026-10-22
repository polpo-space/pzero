package status

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

const (
	GetUserListError = Code(28001)
	StockUnavailable = Code(28002)
	NoMessageError   = Code(28003)
)

func TestError(t *testing.T) {
	err := Error(GetUserListError)
	status := FromError(err)
	assert.Equal(t, GetUserListError, status.Code())
	assert.Equal(t, "get user list error", status.Message())
}

func TestUnknownError(t *testing.T) {
	err := Error(28000)

	status := FromError(err)
	assert.Equal(t, http.StatusInternalServerError, int(status.Code()))
}

func TestWrap(t *testing.T) {
	err := Wrap(GetUserListError, errors.New("connect to db error"))
	status := FromError(err)
	assert.Equal(t, GetUserListError, status.Code())
	assert.Equal(t, "get user list error: connect to db error", status.Error())
	assert.Equal(t, "connect to db error", status.Unwrap().Error())
}

// TestMessageNeverLeaksCause Message() 面向客户端, 任何情况下都不包含底层 cause。
func TestMessageNeverLeaksCause(t *testing.T) {
	cause := errors.New("pq: relation \"users\" does not exist")

	wrapped := FromError(Wrap(GetUserListError, cause))
	assert.Equal(t, "get user list error", wrapped.Message())
	assert.Contains(t, wrapped.Error(), cause.Error())

	internal := FromError(Wrap(http.StatusInternalServerError, cause))
	assert.Equal(t, "internal server error", internal.Message())

	plain := FromError(cause)
	assert.Equal(t, Code(http.StatusInternalServerError), plain.Code())
	assert.Equal(t, "internal server error", plain.Message())
	assert.Equal(t, "internal server error: "+cause.Error(), plain.Error())
}

// TestMessageFallback 无显式提示时依次回退到注册提示、HTTP 标准文案、默认文案。
func TestMessageFallback(t *testing.T) {
	assert.Equal(t, "Not Found", New(http.StatusNotFound, "", nil).Message())
	assert.Equal(t, "internal server error", FromError(Error(NoMessageError)).Message())
	assert.Equal(t, "custom", New(NoMessageError, "custom", nil).Message())
}

// TestGRPCRoundTrip 业务错误经 gRPC status 往返后业务码与提示不丢失, cause 不上线。
func TestGRPCRoundTrip(t *testing.T) {
	err := Wrap(GetUserListError, errors.New("connect to db error"))

	st, ok := grpcstatus.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unknown, st.Code())
	assert.Equal(t, "get user list error", st.Message())

	restored := FromError(st.Err())
	assert.Equal(t, GetUserListError, restored.Code())
	assert.Equal(t, "get user list error", restored.Message())
}

// TestGRPCCodeMapping HTTP 语义的业务码映射到对应 gRPC code, 内置预注册无需再注册。
func TestGRPCCodeMapping(t *testing.T) {
	st, ok := grpcstatus.FromError(Error(http.StatusNotFound))
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())

	assert.Equal(t, Code(http.StatusNotFound), FromError(st.Err()).Code())
}

// TestWithGRPCCode 注册时显式指定的 gRPC code 覆盖默认映射。
func TestWithGRPCCode(t *testing.T) {
	st, ok := grpcstatus.FromError(Error(StockUnavailable))
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
	assert.Equal(t, "stock unavailable", st.Message())

	restored := FromError(st.Err())
	assert.Equal(t, StockUnavailable, restored.Code())
}

// TestFromLegacyGRPCError 不携带 ErrorInfo 的普通 gRPC 错误按 grpc-gateway 规则反向映射。
func TestFromLegacyGRPCError(t *testing.T) {
	cases := map[codes.Code]Code{
		codes.PermissionDenied:   http.StatusForbidden,
		codes.FailedPrecondition: http.StatusBadRequest,
		codes.Aborted:            http.StatusConflict,
		codes.OutOfRange:         http.StatusBadRequest,
		codes.Unknown:            http.StatusInternalServerError,
		codes.DataLoss:           http.StatusInternalServerError,
		codes.Canceled:           499,
	}
	for grpcCode, want := range cases {
		status := FromError(grpcstatus.Error(grpcCode, "legacy"))
		assert.Equal(t, want, status.Code(), grpcCode.String())
		assert.Equal(t, "legacy", status.Message())
	}
}

// TestFromPlainError 普通 error 退化为 500 且保留底层错误供日志使用。
func TestFromPlainError(t *testing.T) {
	cause := errors.New("boom")
	status := FromError(errors.Wrap(cause, "outer"))
	assert.Equal(t, Code(http.StatusInternalServerError), status.Code())
	assert.Equal(t, cause, status.Unwrap())
}

// TestBuiltinHTTPCodes grpcCodes 中的通用 HTTP 码无需注册即可使用, 不再静默变成 500。
func TestBuiltinHTTPCodes(t *testing.T) {
	for code, grpcCode := range grpcCodes {
		status := FromError(Error(code))
		assert.Equal(t, code, status.Code())
		want := http.StatusText(int(code))
		if code == http.StatusInternalServerError {
			want = defaultMessage
		}
		assert.Equal(t, want, status.Message())

		st, ok := grpcstatus.FromError(Error(code))
		assert.True(t, ok)
		assert.Equal(t, grpcCode, st.Code())
	}
}

// TestRegisterOverridesBuiltinHTTPCode 应用注册的提示覆盖内置默认值。
func TestRegisterOverridesBuiltinHTTPCode(t *testing.T) {
	assert.Equal(t, "Conflict", FromError(Error(http.StatusConflict)).Message())
	RegisterWithMessage(http.StatusConflict, "资源冲突")
	t.Cleanup(func() { RegisterWithMessage(http.StatusConflict, http.StatusText(http.StatusConflict)) })

	assert.Equal(t, "资源冲突", FromError(Error(http.StatusConflict)).Message())
}

// TestLookup 按码查询注册信息, 未注册返回 false。
func TestLookup(t *testing.T) {
	registered, ok := Lookup(GetUserListError)
	assert.True(t, ok)
	assert.Equal(t, GetUserListError, registered.Code())
	assert.Equal(t, "get user list error", registered.Message())

	builtin, ok := Lookup(http.StatusNotFound)
	assert.True(t, ok)
	assert.Equal(t, Code(http.StatusNotFound), builtin.Code())

	_, ok = Lookup(28000)
	assert.False(t, ok)
}

// TestOrigin FromError 标明 Status 的来源。
func TestOrigin(t *testing.T) {
	assert.Equal(t, OriginLocal, FromError(Wrap(GetUserListError, errors.New("cause"))).Origin())
	assert.Equal(t, OriginLocal, FromError(fmt.Errorf("outer: %w", Error(GetUserListError))).Origin())

	remote := FromError(grpcstatus.Convert(Error(GetUserListError)).Err())
	assert.Equal(t, OriginRemote, remote.Origin())
	assert.Equal(t, GetUserListError, remote.Code())

	assert.Equal(t, OriginGRPC, FromError(grpcstatus.Error(codes.InvalidArgument, "legacy")).Origin())
	assert.Equal(t, OriginUnknown, FromError(errors.New("boom")).Origin())
}

// TestFromErrorStatusPointer New 返回的 *Status 作为 error 时按本地 Status 还原, 保留 cause 与 extra。
// 之前它被当作 gRPC 错误经 GRPCStatus() 还原, cause 与 extra 丢失。
func TestFromErrorStatusPointer(t *testing.T) {
	cause := errors.New("token exchange failed")
	pointer := New(http.StatusForbidden, "no valid token received", cause)
	pointer.extra = map[string]string{"reason": "expired"}

	status := FromError(fmt.Errorf("callback: %w", pointer))
	assert.Equal(t, OriginLocal, status.Origin())
	assert.Equal(t, Code(http.StatusForbidden), status.Code())
	assert.Equal(t, "no valid token received", status.Message())
	assert.Equal(t, map[string]string{"reason": "expired"}, status.Extra())
	assert.Equal(t, cause, status.Unwrap())
}

func init() {
	RegisterWithMessage(GetUserListError, "get user list error")
	Register(StockUnavailable, WithMessage("stock unavailable"), WithGRPCCode(codes.FailedPrecondition))
	Register(NoMessageError)
}
