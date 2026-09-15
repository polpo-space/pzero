package errcode

import (
	"net/http"

	"github.com/polpo-space/pzero/core/status"
)

// 业务错误码统一定义。
//
// 每个错误码在此处声明的同时即完成注册, 无需再手动调用 status.Register*。
// 新增错误码只需追加一行, 建议按业务模块划分号段, 例如用户模块 10001~10999。
//
// 在 logic 中的用法:
//
//	return nil, status.Error(errcode.NotFound)                          // 使用默认提示
//	return nil, status.ErrorMessage(errcode.InvalidParam, "name 不能为空") // 覆盖提示
//	return nil, status.Wrap(errcode.Internal, err)                      // 底层错误只进日志, 不进 gRPC message
//
// 需要为业务码指定 gRPC code 时(需本仓库发布后的 core/status):
//
//	status.Register(12001, status.WithMessage("stock unavailable"), status.WithGRPCCode(codes.FailedPrecondition))
//
// 该错误可直接作为 gRPC handler 的返回值: gRPC code 由 core/status 按语义映射(可用 WithGRPCCode 覆盖),
// message 只传 Message(), 业务码随 ErrorInfo detail 传输, 调用方通过 status.FromError 即可还原。
var (
	InvalidParam = register(http.StatusBadRequest, "invalid parameter")
	Unauthorized = register(http.StatusUnauthorized, "unauthorized")
	Forbidden    = register(http.StatusForbidden, "forbidden")
	NotFound     = register(http.StatusNotFound, "not found")
	Internal     = register(http.StatusInternalServerError, "internal server error")
)

// register 注册错误码及其默认提示并返回该错误码, 保证定义与注册不会脱节。
func register(code status.Code, message string) status.Code {
	status.RegisterWithMessage(code, message)
	return code
}
