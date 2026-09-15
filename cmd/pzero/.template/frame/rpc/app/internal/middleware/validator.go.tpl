package middleware

import (
	"context"

	"buf.build/go/protovalidate"
	"github.com/pkg/errors"
	"github.com/polpo-space/pzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"{{ .Module }}/internal/errcode"
)

type Validator struct {
	v protovalidate.Validator
}

func NewValidator() *Validator {
	v, err := protovalidate.New()
	logx.Must(err)
	return &Validator{v: v}
}

func (v *Validator) UnaryServerMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		switch req.(type) {
		case proto.Message:
			if err := v.v.Validate(req.(proto.Message)); err != nil {
				// 校验失败统一为 InvalidParam 业务码, 由 core/status 映射为 gRPC InvalidArgument
				var valErr *protovalidate.ValidationError
				if ok := errors.As(err, &valErr); ok && len(valErr.ToProto().GetViolations()) > 0 {
					return nil, status.ErrorMessage(errcode.InvalidParam, valErr.ToProto().GetViolations()[0].GetMessage())
				}
				return nil, status.ErrorMessage(errcode.InvalidParam, err.Error())
			}
		}
		return handler(ctx, req)
	}
}
