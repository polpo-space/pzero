package status

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestErrorDetailsRoundTrip details 随 gRPC 往返, 排在 pzero ErrorInfo 之后, 调用方经 Details() 读取。
func TestErrorDetailsRoundTrip(t *testing.T) {
	resource := &errdetails.ResourceInfo{ResourceType: "user", ResourceName: "42"}
	err := ErrorDetails(StockUnavailable, resource)

	local := FromError(err)
	assert.Equal(t, OriginLocal, local.Origin())
	require.Len(t, local.Details(), 1)
	assert.True(t, proto.Equal(resource, local.Details()[0]))

	wire := grpcstatus.Convert(err)
	assert.Equal(t, codes.FailedPrecondition, wire.Code())
	require.Len(t, wire.Details(), 2)
	info, ok := wire.Details()[0].(*errdetails.ErrorInfo)
	require.True(t, ok)
	assert.Equal(t, errorInfoDomain, info.GetDomain())
	assert.True(t, proto.Equal(resource, wire.Details()[1].(proto.Message)))

	remote := FromError(wire.Err())
	assert.Equal(t, OriginRemote, remote.Origin())
	assert.Equal(t, StockUnavailable, remote.Code())
	require.Len(t, remote.Details(), 1)
	assert.True(t, proto.Equal(resource, remote.Details()[0]))
}

// TestRelayKeepsSingleErrorInfo 中间服务原样或按 *FromError 转发下游错误, 线上只有一个 pzero ErrorInfo 且 details 保留。
func TestRelayKeepsSingleErrorInfo(t *testing.T) {
	resource := &errdetails.ResourceInfo{ResourceType: "user"}
	fromC := grpcstatus.Convert(ErrorDetails(StockUnavailable, resource)).Err()

	for name, relayed := range map[string]error{
		"as is":       fromC,
		"*FromError":  *FromError(fromC),
		"wrapped %w":  fmt.Errorf("relay: %w", *FromError(fromC)),
		"interceptor": serverError(context.Background(), "/m", fmt.Errorf("relay: %w", fromC)),
	} {
		t.Run(name, func(t *testing.T) {
			wire := grpcstatus.Convert(serverError(context.Background(), "/m", relayed))
			infos := 0
			for _, detail := range wire.Details() {
				if info, ok := detail.(*errdetails.ErrorInfo); ok && info.GetDomain() == errorInfoDomain {
					infos++
				}
			}
			assert.Equal(t, 1, infos)
			_, isInfo := wire.Details()[0].(*errdetails.ErrorInfo)
			assert.True(t, isInfo, "ErrorInfo must come first")
			restored := FromError(wire.Err())
			assert.Equal(t, StockUnavailable, restored.Code())
			require.Len(t, restored.Details(), 1)
			assert.True(t, proto.Equal(resource, restored.Details()[0]))
		})
	}
}

// TestDetailsKeepsForeignErrorInfoAndSkipsUnknownTypes 只排除 pzero 自己的 ErrorInfo, 无法解析的 detail 跳过。
func TestDetailsKeepsForeignErrorInfoAndSkipsUnknownTypes(t *testing.T) {
	pzeroInfo, err := anypb.New(&errdetails.ErrorInfo{Domain: errorInfoDomain, Reason: "28001"})
	require.NoError(t, err)
	foreignInfo, err := anypb.New(&errdetails.ErrorInfo{Domain: "other", Reason: "x"})
	require.NoError(t, err)
	unknown := &anypb.Any{TypeUrl: "type.googleapis.com/unknown.Type", Value: []byte{1}}

	wire := grpcstatus.FromProto(&spb.Status{
		Code:    int32(codes.FailedPrecondition),
		Message: "m",
		Details: []*anypb.Any{pzeroInfo, foreignInfo, unknown},
	})
	restored := FromError(wire.Err())
	assert.Equal(t, GetUserListError, restored.Code())
	require.Len(t, restored.Details(), 1)
	assert.Equal(t, "other", restored.Details()[0].(*errdetails.ErrorInfo).GetDomain())
}

// TestDetailsAreCopied 修改 Details() 的返回值不影响 Status。
func TestDetailsAreCopied(t *testing.T) {
	status := FromError(ErrorDetails(StockUnavailable, &errdetails.ResourceInfo{ResourceType: "user"}))
	status.Details()[0] = nil
	assert.NotNil(t, status.Details()[0])
}

// TestStatusStaysComparable 新字段不破坏 Status 的可比较性与 errors.Is。
func TestStatusStaysComparable(t *testing.T) {
	assert.True(t, errors.Is(Error(GetUserListError), Error(GetUserListError)))
	assert.NotPanics(t, func() {
		_ = ErrorDetails(StockUnavailable, &errdetails.ResourceInfo{}) == Error(StockUnavailable)
	})
}

// TestUnregisteredErrorDetails 未注册的 code 与 Error 一样回退为 500。
func TestUnregisteredErrorDetails(t *testing.T) {
	assert.Equal(t, Code(http.StatusInternalServerError), FromError(ErrorDetails(28000, &errdetails.ResourceInfo{})).Code())
}

// detailWrapper 模拟调用方自己实现 GRPCStatus 的包装器(例如过渡期的 details helper)。
type detailWrapper struct {
	base   error
	detail proto.Message
}

func (w detailWrapper) Error() string { return w.base.Error() }
func (w detailWrapper) Unwrap() error { return w.base }
func (w detailWrapper) GRPCStatus() *grpcstatus.Status {
	st, _ := grpcstatus.Convert(w.base).WithDetails(w.detail.(*errdetails.ResourceInfo))
	return st
}

func startFailServer(t *testing.T, err error) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer(grpc.UnaryInterceptor(UnaryServerInterceptor))
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: "pzero.status.Fail",
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{{
			MethodName: "Call",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				in := new(wrapperspb.StringValue)
				if decodeErr := dec(in); decodeErr != nil {
					return nil, decodeErr
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/pzero.status.Fail/Call"}
				return interceptor(ctx, in, info, func(context.Context, any) (any, error) { return nil, err })
			},
		}},
	}, struct{}{})
	go func() { _ = server.Serve(lis) }()
	t.Cleanup(server.Stop)

	conn, dialErr := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, dialErr)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// TestServerInterceptor 经真实 gRPC server 验证拦截器的输出。
func TestServerInterceptor(t *testing.T) {
	resource := &errdetails.ResourceInfo{ResourceType: "user"}
	tests := []struct {
		name        string
		err         error
		wantCode    codes.Code
		wantMessage string
		wantBiz     Code
		wantDetails int
	}{
		{"wrapped status keeps message", fmt.Errorf("load user 42: %w", Wrap(GetUserListError, errors.New(leakCause))), codes.Unknown, "get user list error", GetUserListError, 0},
		{"wrapped status keeps details", fmt.Errorf("x: %w", ErrorDetails(StockUnavailable, resource)), codes.FailedPrecondition, "stock unavailable", StockUnavailable, 1},
		{"outer GRPCStatus wrapper keeps its details", fmt.Errorf("x: %w", detailWrapper{base: Error(StockUnavailable), detail: resource}), codes.FailedPrecondition, "stock unavailable", StockUnavailable, 1},
		{"status pointer", New(http.StatusForbidden, "no valid token received", errors.New(leakCause)), codes.PermissionDenied, "no valid token received", http.StatusForbidden, 0},
		{"wrapped raw grpc error drops prefix", fmt.Errorf("x: %w", grpcstatus.Error(codes.InvalidArgument, "phone is required")), codes.InvalidArgument, "phone is required", http.StatusBadRequest, 0},
		{"wrapped deadline", fmt.Errorf("query user 42 on host db-1: %w", context.DeadlineExceeded), codes.DeadlineExceeded, context.DeadlineExceeded.Error(), http.StatusGatewayTimeout, 0},
		{"wrapped cancel", fmt.Errorf("query user 42: %w", context.Canceled), codes.Canceled, context.Canceled.Error(), 499, 0},
		{"plain error", errors.New(leakCause), codes.Unknown, defaultMessage, http.StatusInternalServerError, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := startFailServer(t, tt.err)
			err := conn.Invoke(context.Background(), "/pzero.status.Fail/Call", wrapperspb.String("x"), &wrapperspb.StringValue{})
			require.Error(t, err)

			st := grpcstatus.Convert(err)
			assert.Equal(t, tt.wantCode, st.Code())
			assert.Equal(t, tt.wantMessage, st.Message())
			assert.NotContains(t, st.Message(), leakCause)

			restored := FromError(err)
			assert.Equal(t, tt.wantBiz, restored.Code())
			assert.Len(t, restored.Details(), tt.wantDetails)
		})
	}
}

type contextStream struct {
	grpc.ServerStream
}

func (contextStream) Context() context.Context { return context.Background() }

// TestStreamServerInterceptor 流式版本与 unary 输出一致。
func TestStreamServerInterceptor(t *testing.T) {
	info := &grpc.StreamServerInfo{FullMethod: "/pzero.status.Fail/Stream"}
	err := StreamServerInterceptor(nil, contextStream{}, info, func(any, grpc.ServerStream) error {
		return fmt.Errorf("load user 42: %w", Wrap(GetUserListError, errors.New(leakCause)))
	})
	st := grpcstatus.Convert(err)
	assert.Equal(t, "get user list error", st.Message())
	assert.Equal(t, GetUserListError, FromError(err).Code())

	assert.NoError(t, StreamServerInterceptor(nil, contextStream{}, info, func(any, grpc.ServerStream) error { return nil }))
}
