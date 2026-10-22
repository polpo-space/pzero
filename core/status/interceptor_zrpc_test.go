package status

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestServerErrorKeepsCauseChain 拦截器返回的错误仍能沿 Unwrap 识别 cause, go-zero 熔断据此判断失败。
func TestServerErrorKeepsCauseChain(t *testing.T) {
	timeout := serverError(context.Background(), "/m", fmt.Errorf("x: %w", Wrap(GetUserListError, context.DeadlineExceeded)))
	assert.True(t, errorx.In(timeout, context.DeadlineExceeded))
	assert.Equal(t, "get user list error", timeout.Error())

	unavailable := serverError(context.Background(), "/m", fmt.Errorf("call stock: %w", breaker.ErrServiceUnavailable))
	assert.True(t, errors.Is(unavailable, breaker.ErrServiceUnavailable))
	assert.Equal(t, defaultMessage, unavailable.Error())
}

// TestFromErrorFiltersEveryPzeroErrorInfo 重复或 Reason 非法的 pzero ErrorInfo 都不进 Details(), 转发后线上只剩一个。
func TestFromErrorFiltersEveryPzeroErrorInfo(t *testing.T) {
	first, err := anypb.New(&errdetails.ErrorInfo{Domain: errorInfoDomain, Reason: "28001"})
	require.NoError(t, err)
	duplicate, err := anypb.New(&errdetails.ErrorInfo{Domain: errorInfoDomain, Reason: "28002"})
	require.NoError(t, err)
	invalid, err := anypb.New(&errdetails.ErrorInfo{Domain: errorInfoDomain, Reason: "not-a-code"})
	require.NoError(t, err)
	resource, err := anypb.New(&errdetails.ResourceInfo{ResourceType: "user"})
	require.NoError(t, err)

	wire := grpcstatus.FromProto(&spb.Status{
		Code:    int32(codes.FailedPrecondition),
		Message: "m",
		Details: []*anypb.Any{invalid, first, duplicate, resource},
	})
	restored := FromError(wire.Err())
	assert.Equal(t, GetUserListError, restored.Code())
	require.Len(t, restored.Details(), 1)
	assert.Equal(t, "user", restored.Details()[0].(*errdetails.ResourceInfo).GetResourceType())

	infos := 0
	for _, detail := range grpcstatus.Convert(*restored).Details() {
		if info, ok := detail.(*errdetails.ErrorInfo); ok && info.GetDomain() == errorInfoDomain {
			infos++
		}
	}
	assert.Equal(t, 1, infos)

	onlyInvalid := FromError(grpcstatus.FromProto(&spb.Status{Code: int32(codes.InvalidArgument), Details: []*anypb.Any{invalid}}).Err())
	assert.Equal(t, OriginGRPC, onlyInvalid.Origin())
	assert.Empty(t, onlyInvalid.Details())
}

// startZrpcServer 启动带 go-zero 默认中间件(含熔断)的 zrpc 服务, handler 返回 handlerErr。
func startZrpcServer(t *testing.T, service string, withInterceptor bool, handlerErr error) *grpc.ClientConn {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	require.NoError(t, lis.Close())

	var c zrpc.RpcServerConf
	require.NoError(t, conf.FillDefault(&c))
	c.Name = service
	c.ListenOn = addr
	c.Health = false

	server := zrpc.MustNewServer(c, func(s *grpc.Server) {
		s.RegisterService(&grpc.ServiceDesc{
			ServiceName: service,
			HandlerType: (*any)(nil),
			Methods: []grpc.MethodDesc{{
				MethodName: "Call",
				Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
					in := new(wrapperspb.StringValue)
					if decodeErr := dec(in); decodeErr != nil {
						return nil, decodeErr
					}
					info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + service + "/Call"}
					return interceptor(ctx, in, info, func(context.Context, any) (any, error) { return nil, handlerErr })
				},
			}},
		}, struct{}{})
	})
	if withInterceptor {
		server.AddUnaryInterceptors(UnaryServerInterceptor)
	}
	go server.Start()
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	require.Eventually(t, func() bool {
		dial, dialErr := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if dialErr != nil {
			return false
		}
		_ = dial.Close()
		return true
	}, 5*time.Second, 20*time.Millisecond)
	return conn
}

func callZrpc(conn *grpc.ClientConn, service string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return conn.Invoke(ctx, "/"+service+"/Call", wrapperspb.String("x"), &wrapperspb.StringValue{})
}

// TestInterceptorKeepsZrpcBreakerSemantics 与 go-zero 默认中间件组合时, 熔断的错误转换与失败计数和不启用拦截器时一致。
func TestInterceptorKeepsZrpcBreakerSemantics(t *testing.T) {
	t.Run("wrapped breaker.ErrServiceUnavailable stays Unavailable", func(t *testing.T) {
		for _, withInterceptor := range []bool{false, true} {
			service := fmt.Sprintf("pzero.status.Unavailable%v", withInterceptor)
			conn := startZrpcServer(t, service, withInterceptor, fmt.Errorf("call stock 42: %w", breaker.ErrServiceUnavailable))
			err := callZrpc(conn, service)
			assert.Equal(t, codes.Unavailable, grpcstatus.Code(err), "interceptor=%v", withInterceptor)
			if withInterceptor {
				assert.NotContains(t, grpcstatus.Convert(err).Message(), "call stock 42")
			}
		}
	})

	t.Run("business error caused by timeout still trips the breaker", func(t *testing.T) {
		for _, withInterceptor := range []bool{false, true} {
			service := fmt.Sprintf("pzero.status.Timeout%v", withInterceptor)
			conn := startZrpcServer(t, service, withInterceptor, fmt.Errorf("x: %w", Wrap(GetUserListError, context.DeadlineExceeded)))
			rejected := 0
			for range 200 {
				if grpcstatus.Code(callZrpc(conn, service)) == codes.Unavailable {
					rejected++
				}
			}
			assert.Positive(t, rejected, "interceptor=%v: breaker must count the timeout cause as a failure", withInterceptor)
		}
	})
}
