package status

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const bufSize = 1024 * 1024

const leakCause = `pq: relation "users" does not exist`

type echoServer interface {
	Call(context.Context, *wrapperspb.StringValue) (*wrapperspb.StringValue, error)
}

type echoImpl struct{}

func (echoImpl) Call(_ context.Context, in *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	cause := errors.New(leakCause)
	switch in.GetValue() {
	case "wrap-internal":
		return nil, Wrap(http.StatusInternalServerError, cause)
	case "business":
		return nil, Wrap(GetUserListError, cause)
	case "custom-grpc":
		return nil, Wrap(StockUnavailable, cause)
	default:
		return nil, Error(http.StatusInternalServerError)
	}
}

func echoHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(wrapperspb.StringValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(echoServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/pzero.status.Echo/Call"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(echoServer).Call(ctx, req.(*wrapperspb.StringValue))
	}
	return interceptor(ctx, in, info, handler)
}

func startBufconn(t *testing.T) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "pzero.status.Echo",
		HandlerType: (*echoServer)(nil),
		Methods: []grpc.MethodDesc{{
			MethodName: "Call",
			Handler:    echoHandler,
		}},
	}, echoImpl{})
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func invoke(t *testing.T, conn *grpc.ClientConn, kind string) error {
	t.Helper()
	err := conn.Invoke(context.Background(), "/pzero.status.Echo/Call", wrapperspb.String(kind), &wrapperspb.StringValue{})
	require.Error(t, err)
	return err
}

// TestBufconnRoundTrip 走真实 gRPC server/client, 验证业务码、ErrorInfo、gRPC code 与 cause 不上线。
func TestBufconnRoundTrip(t *testing.T) {
	conn := startBufconn(t)

	t.Run("wrap-internal", func(t *testing.T) {
		err := invoke(t, conn, "wrap-internal")
		st, ok := grpcstatus.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "internal server error", st.Message())
		assert.False(t, strings.Contains(st.Message(), leakCause))

		restored := FromError(err)
		assert.Equal(t, Code(http.StatusInternalServerError), restored.Code())
		assert.Equal(t, "internal server error", restored.Message())
		assert.False(t, strings.Contains(restored.Error(), leakCause))
	})

	t.Run("business", func(t *testing.T) {
		err := invoke(t, conn, "business")
		st, ok := grpcstatus.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unknown, st.Code())
		assert.Equal(t, "get user list error", st.Message())
		assert.False(t, strings.Contains(st.Message(), leakCause))

		require.Len(t, st.Details(), 1)
		info, ok := st.Details()[0].(*errdetails.ErrorInfo)
		require.True(t, ok)
		assert.Equal(t, errorInfoDomain, info.GetDomain())
		assert.Equal(t, "28001", info.GetReason())

		restored := FromError(err)
		assert.Equal(t, GetUserListError, restored.Code())
		assert.Equal(t, "get user list error", restored.Message())
	})

	t.Run("custom-grpc", func(t *testing.T) {
		err := invoke(t, conn, "custom-grpc")
		st, ok := grpcstatus.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Equal(t, "stock unavailable", st.Message())
		assert.False(t, strings.Contains(st.Message(), leakCause))

		restored := FromError(err)
		assert.Equal(t, StockUnavailable, restored.Code())
	})
}
