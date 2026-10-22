package status

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// UnaryServerInterceptor 统一 RPC handler 返回错误的 gRPC 输出, 需要的服务通过 AddUnaryInterceptors 自行启用。
//
//   - 带 GRPCStatus 的错误(Status、*Status、下游透传的 gRPC 错误)按其原本的 status 返回。
//     grpc-go 对被 fmt.Errorf("%w") 等包装过的错误会把 message 换成 err.Error(), 这里取包装前的 status,
//     message 与 details 不受包装影响。errors.As 取最外层的提供者, 包装器自带的 details 得以保留。
//   - context 取消或超时按 Canceled / DeadlineExceeded 返回, message 固定, 不带包装文本。
//   - 其他错误保持 grpc-go 默认的 Unknown(熔断计数不变), message 换成默认文案, 原始错误在此记一条 ERROR。
//
// 返回的错误保留原错误链, 外层的 go-zero 熔断拦截器判断失败与转换 breaker.ErrServiceUnavailable 的行为不变。
//
// go-zero 内置的 recover 拦截器在外层, panic 不经过这里。
func UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return resp, serverError(ctx, info.FullMethod, err)
	}
	return resp, nil
}

// StreamServerInterceptor 是 UnaryServerInterceptor 的流式版本。
func StreamServerInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := handler(srv, stream); err != nil {
		return serverError(stream.Context(), info.FullMethod, err)
	}
	return nil
}

func serverError(ctx context.Context, method string, err error) error {
	if st := grpcStatusOf(err); st != nil {
		return &serverStatusError{status: st, cause: err}
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return &serverStatusError{status: grpcstatus.New(codes.DeadlineExceeded, context.DeadlineExceeded.Error()), cause: err}
	case errors.Is(err, context.Canceled):
		return &serverStatusError{status: grpcstatus.New(codes.Canceled, context.Canceled.Error()), cause: err}
	}
	logx.WithContext(ctx).Errorf("%s: %v", method, err)
	return &serverStatusError{status: grpcstatus.New(codes.Unknown, defaultMessage), cause: err}
}

// serverStatusError 携带要上线的 status, 同时保留原错误链。
// grpc-go 对自身实现 GRPCStatus 的错误直接使用该 status, 不会把 message 换成 Error();
// 外层拦截器(如 go-zero 熔断按 context.DeadlineExceeded、breaker.ErrServiceUnavailable 判断)仍能沿 Unwrap 识别 cause。
// Error() 返回上线的 message, 外层用 Error() 构造新 status 时也不会带出 cause。
type serverStatusError struct {
	status *grpcstatus.Status
	cause  error
}

func (e *serverStatusError) Error() string { return e.status.Message() }

func (e *serverStatusError) Unwrap() error { return e.cause }

func (e *serverStatusError) GRPCStatus() *grpcstatus.Status { return e.status }

// grpcStatusOf 返回 err 链上第一个 GRPCStatus 提供者的 status; 没有、为 nil 或为 OK 时返回 nil。
func grpcStatusOf(err error) *grpcstatus.Status {
	var provider interface{ GRPCStatus() *grpcstatus.Status }
	if !errors.As(err, &provider) {
		return nil
	}
	st := provider.GRPCStatus()
	if st == nil || st.Code() == codes.OK {
		return nil
	}
	return st
}
