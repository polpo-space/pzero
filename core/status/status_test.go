package status

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

const (
	GetUserListError = Code(28001)
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

// TestGRPCRoundTrip 业务错误经 gRPC status 往返后业务码与提示不丢失。
func TestGRPCRoundTrip(t *testing.T) {
	err := Wrap(GetUserListError, errors.New("connect to db error"))

	st, ok := grpcstatus.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unknown, st.Code())
	assert.Equal(t, "get user list error: connect to db error", st.Message())

	restored := FromError(st.Err())
	assert.Equal(t, GetUserListError, restored.Code())
	assert.Equal(t, "get user list error: connect to db error", restored.Error())
}

// TestGRPCCodeMapping HTTP 语义的业务码映射到对应 gRPC code。
func TestGRPCCodeMapping(t *testing.T) {
	RegisterWithMessage(http.StatusNotFound, "not found")

	st, ok := grpcstatus.FromError(Error(http.StatusNotFound))
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())

	assert.Equal(t, Code(http.StatusNotFound), FromError(st.Err()).Code())
}

// TestFromPlainGRPCError 不携带 ErrorInfo 的普通 gRPC 错误按 code 反向映射。
func TestFromPlainGRPCError(t *testing.T) {
	status := FromError(grpcstatus.Error(codes.PermissionDenied, "no permission"))
	assert.Equal(t, Code(http.StatusForbidden), status.Code())
	assert.Equal(t, "no permission", status.Message())

	status = FromError(grpcstatus.Error(codes.Aborted, "aborted"))
	assert.Equal(t, Code(http.StatusInternalServerError), status.Code())
}

// TestFromPlainError 普通 error 仍退化为 500 且保留底层错误。
func TestFromPlainError(t *testing.T) {
	cause := errors.New("boom")
	status := FromError(errors.Wrap(cause, "outer"))
	assert.Equal(t, Code(http.StatusInternalServerError), status.Code())
	assert.Equal(t, cause, status.Unwrap())
}

func init() {
	RegisterWithMessage(GetUserListError, "get user list error")
}
