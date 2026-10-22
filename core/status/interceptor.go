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
		return st.Err()
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return grpcstatus.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	case errors.Is(err, context.Canceled):
		return grpcstatus.Error(codes.Canceled, context.Canceled.Error())
	}
	logx.WithContext(ctx).Errorf("%s: %v", method, err)
	return grpcstatus.Error(codes.Unknown, defaultMessage)
}

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
