---
title: 错误处理指南
icon: /icons/eos-icons-api.svg
star: true
order: 0.3
---

# 错误处理指南

这份指南是 API / RPC / BFF 写错误时的单一事实来源。后续改业务逻辑、加接口、接 RPC，都按这里做，不要发明第二套错误体系。

## 一句话

业务错误是一个整数。发出方用 `internal/errcode` 注册数字和对外文案，logic 只返回 `status.Error` / `status.Wrap` / `status.ErrorMessage`。HTTP 的 `msg` 和 gRPC message 只发 `Message()`，底层 cause 只留在本进程日志里。

## 必须先分清的两个字符串

| 方法 | 给谁看 | 含不含 cause |
| --- | --- | --- |
| `Error()` | 日志、Go error chain | 含。`Wrap(Internal, dbErr)` 会拼出 `internal server error: pq: ...` |
| `Message()` | 客户端、HTTP `msg`、gRPC message | 不含。500 没有显式文案时固定为 `internal server error` |

脚手架里的 API 响应中间件已经是 `Msg: fromError.Message()`。不要改回 `Error()`。

`status.Wrap` 返回的就是 `Status` 本身。不要再套 `errors.WithStack` 或 `fmt.Errorf("%w")` 之后从 gRPC handler 返回：grpc-go 对 wrap 过的 status 会把 message 改成 `err.Error()`，cause 立刻上线。

## 数据怎么走

```text
logic
  return status.Error(errcode.X)          // 或 Wrap / ErrorMessage
        │
        ├─ API: ErrorMiddleware
        │     logx 打 Error()（含 cause）
        │     body = { code: X, msg: Message(), data: null }   HTTP 仍是 200
        │
        └─ RPC: Status.GRPCStatus()
              gRPC code ← 映射表或 WithGRPCCode
              message   ← Message()
              details   ← ErrorInfo{ domain: "pzero", reason: "X" }
                    │
                    └─ BFF: return nil, err          // 默认透传
                          status.FromError(err)      // 还原 X 和 Message()
```

BFF 不需要在自己的 `errcode` 里注册 RPC 的码，也能从 `FromError` 读到数字和文案。

## 决策树

拿到一个 `err`，按顺序只走一条：

1. **这是当前服务自己的业务失败**（找不到、状态不允许、参数语义错）→ `return nil, status.Error(errcode.Xxx)`。需要临时改文案 → `ErrorMessage`。
2. **这是基础设施失败**（DB、缓存、SDK、未知 err）→ `return nil, status.Wrap(errcode.Internal, err)`。
3. **这是 model 的 `ErrNotFound`，且对外就是 404** → `return nil, status.Error(errcode.NotFound)`。
4. **这是调用下游 RPC 的错误**
   - 前端就该看到下游的码和文案 → `return nil, err`（透传）。
   - 本接口要按「库存不足 / 已有待处理单」分支 → `status.FromError(err).Code() == status.Code(pb.Xxx)`，数字来自 contracts enum，不要比 message 字符串。
   - 本接口要换一个对公网的码或文案 → `status.ErrorMessage(errcode.Y, "对公网文案")`。
5. **不要** `errors.New("用户不存在")`，**不要** `grpc/status.Error(codes.NotFound, ...)`，**不要**手写 `codes → http` switch。

未注册的 code 传给 `status.Error` 会静默变成 500。所以数字必须来自 `register(...)` 的返回值，不要写裸 `status.Code(10001)`。

## 注册错误码

`pzero new --frame api|rpc` 会生成 `internal/errcode/errcode.go`。定义即注册：

```go
var (
	InvalidParam = register(http.StatusBadRequest, "invalid parameter")
	Unauthorized = register(http.StatusUnauthorized, "unauthorized")
	Forbidden    = register(http.StatusForbidden, "forbidden")
	NotFound     = register(http.StatusNotFound, "not found")
	Internal     = register(http.StatusInternalServerError, "internal server error")
)

func register(code status.Code, message string) status.Code {
	status.RegisterWithMessage(code, message)
	return code
}
```

本地通用码用 HTTP 语义（400/401/403/404/500）。本服务特有的、调用方必须区分的码，用业务号段，例如用户 10001–10999。

需要让不懂 `core/status` 的 gRPC 客户端看到 canonical code 时，在**已升级到本仓库 `core/status`** 的服务里写：

```go
status.Register(
	12001,
	status.WithMessage("stock unavailable"),
	status.WithGRPCCode(codes.FailedPrecondition),
)
```

脚手架生成的 `register()` 故意只调 `RegisterWithMessage`，保证对当前已发布的 `core/status` 也能编译。

## 场景一：纯 API（用户查询 / 注册）

```go
func (l *GetUser) GetUser(req *types.GetUserRequest) (*types.GetUserResponse, error) {
	user, err := l.svcCtx.Model.User.FindOne(l.ctx, req.Id)
	if errors.Is(err, usermodel.ErrNotFound) {
		return nil, status.Error(errcode.NotFound)
	}
	if err != nil {
		return nil, status.Wrap(errcode.Internal, err)
	}
	if user.Disabled {
		return nil, status.ErrorMessage(errcode.UserDisabled, "user is disabled")
	}
	return &types.GetUserResponse{Id: user.Id, Name: user.Name}, nil
}

func (l *Register) Register(req *types.RegisterRequest) (*types.RegisterResponse, error) {
	_, err := l.svcCtx.Model.User.FindOneByUsername(l.ctx, req.Username)
	if err == nil {
		return nil, status.Error(errcode.UserAlreadyExists) // 例如 10002
	}
	if !errors.Is(err, usermodel.ErrNotFound) {
		return nil, status.Wrap(errcode.Internal, err)
	}
	id, err := l.svcCtx.Model.User.Insert(l.ctx, &usermodel.User{Username: req.Username})
	if err != nil {
		return nil, status.Wrap(errcode.Internal, err)
	}
	return &types.RegisterResponse{Id: id}, nil
}
```

客户端看到的是：

```json
{ "code": 404, "msg": "not found", "data": null }
```

或 `{ "code": 10002, "msg": "user already exists", "data": null }`。HTTP 状态码仍是 200，业务状态只看 body 里的 `code`。

`ErrNotFound` 用 `errors.Is`（`github.com/pkg/errors`），不要 `==`。

## 场景二：纯 RPC（同一套码，直接当 gRPC error 返回）

```go
func (l *GetUser) GetUser(in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := l.svcCtx.Model.User.FindOne(l.ctx, in.GetId())
	if errors.Is(err, usermodel.ErrNotFound) {
		return nil, status.Error(errcode.UserNotFound) // 例如 10001
	}
	if err != nil {
		return nil, status.Wrap(errcode.Internal, err)
	}
	return &userv1.GetUserResponse{Id: user.Id, Name: user.Name}, nil
}
```

不要写 `grpcstatus.Error(codes.NotFound, "user not found")`。`Status` 已经实现 `GRPCStatus()`：

- 10001 默认 gRPC `Unknown`，业务码在 `ErrorInfo.reason`
- 404 映射 `NotFound`，500 映射 `Internal`
- message 是 `Message()`，`Wrap` 的 SQL 文案不会出进程
- 服务端 zrpc 日志仍会打完整 `Error()`

校验失败走脚手架的 validator 中间件，已经是 `status.ErrorMessage(errcode.InvalidParam, ...)`，映射为 gRPC `InvalidArgument`。

## 场景三：BFF 调 RPC

### 透传（默认，也是最常见的）

```go
resp, err := l.svcCtx.UserRpc.GetUser(l.ctx, &userv1.GetUserRequest{Id: req.Id})
if err != nil {
	return nil, err
}
```

前端拿到的 `code` 就是 RPC 发出的 10001，`msg` 是 RPC 注册的默认文案。BFF 不用注册 10001。

### 按业务码分支（本接口语义和下游不完全一样）

数字必须来自双方都能 import 的 contracts enum，不能从 BFF 的 `internal/errcode` 抄一份：

```go
st := status.FromError(err)
if st.Code() == status.Code(orderv1.OrderError_ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE) {
	return &types.CreateOrderResponse{SoldOut: true}, nil
}
return nil, err
```

跨进程之后本地 `errors.Is` 对不上 RPC 进程里的 sentinel。只比数字。

### 改写对公网契约（换码或换文案）

```go
if status.FromError(err).Code() == status.Code(nfcv1.NfcError_NFC_ERROR_TAG_NOT_FOUND) {
	return nil, status.ErrorMessage(errcode.NotFound, "内容不存在或不可用")
}
```

对公网隐藏内部细节时才改写。不要为了「统一成 HTTP 404」把所有 RPC 错误先 `mapGrpcError` 一遍——`FromError` 已经带了反向映射。

### 不要做的 BFF 写法

```go
// 错：比 message 字符串，文案一改分支就断
if s.Code() == codes.FailedPrecondition && s.Message() == "preproduction_stock_unavailable" { ... }

// 错：太粗，分不出设备不存在还是门店不存在
if grpcstatus.Code(err) == codes.NotFound { ... }

// 错：每个 BFF 抄一份 codes → http
switch grpcstatus.Code(err) {
case codes.NotFound:
	return jzerostatus.ErrorMessage(http.StatusNotFound, ...)
}
```

## 跨服务共享码：放 contracts 的 proto enum

BFF 必须区分的错误，数字的唯一出处是 `contracts/proto/<svc>/v1/error_code.proto`，不是服务的 `internal` 包（`internal` 跨模块不可 import），也不是 `pkg/` 里的字符串常量。

```protobuf
// contracts/proto/order/v1/error_code.proto
syntax = "proto3";
package order.v1;
option go_package = "github.com/org/repo/contracts/gen/order/v1";

// Range: 12000-12999. 永不改号、不复用。
enum OrderError {
  ORDER_ERROR_UNSPECIFIED = 0;
  ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE = 12001;
  ORDER_ERROR_PREPRODUCTION_PENDING_ORDER_EXISTS = 12002;
}
```

约束：

- `contracts` 模块不依赖 `pzero`。这里只有数字，没有文案，没有 `register()`。
- `contracts/gen` 若 `buf.gen.yaml` 开了 `clean: true`，不要把手写 Go 放进 `gen/`。
- 文案和注册留在发出方 `internal/errcode`：`register(status.Code(orderv1.OrderError_...), "预生产库存不足")`。
- 每服务一个千位号段。protoc 只保证单个 enum 内不重复，跨服务号段用 contracts 测试守。
- 通用 400/401/403/404/500 不进契约。
- 退役用 `reserved`，不要删号、改号。

完整步骤（enum、注册、BFF 比较、号段测试）见 [Shared Error Codes](https://github.com/polpo-space/pzero/blob/main/skills/pzero-skills/references/rpc-patterns/error-codes.md)。

## 映射表（不要在业务里再抄）

HTTP 语义码 → gRPC：

| 业务码 | gRPC |
| --- | --- |
| 400 | InvalidArgument |
| 401 | Unauthenticated |
| 403 | PermissionDenied |
| 404 | NotFound |
| 409 | AlreadyExists |
| 429 | ResourceExhausted |
| 500 | Internal |
| 501 | Unimplemented |
| 503 | Unavailable |
| 504 | DeadlineExceeded |
| 其他业务号段 | Unknown（可用 `WithGRPCCode` 覆盖） |

无 `ErrorInfo` 的旧 gRPC 错误 → HTTP（grpc-gateway 规则）：

| gRPC | 业务码 |
| --- | --- |
| InvalidArgument / FailedPrecondition / OutOfRange | 400 |
| Unauthenticated | 401 |
| PermissionDenied | 403 |
| NotFound | 404 |
| AlreadyExists / Aborted | 409 |
| ResourceExhausted | 429 |
| Canceled | 499 |
| Unimplemented | 501 |
| Unavailable | 503 |
| DeadlineExceeded | 504 |
| Unknown / Internal / DataLoss / 未列出 | 500 |

go-zero zrpc 熔断只把 `DeadlineExceeded` / `Internal` / `Unavailable` / `DataLoss` / `Unimplemented` / `ResourceExhausted` 计为失败。业务码默认 `Unknown`，以及 `FailedPrecondition`，不会踩熔断。

## 正确 vs 错误

| 场景 | 正确 | 错误 |
| --- | --- | --- |
| 用户不存在 | `status.Error(errcode.UserNotFound)` | `errors.New("用户不存在")` |
| DB 挂了 | `status.Wrap(errcode.Internal, err)` | `return nil, err` 把 `pq:` 丢给客户端 |
| RPC 找不到 | `status.Error(errcode.X)` | `grpcstatus.Error(codes.NotFound, ...)` |
| BFF 调 RPC | `return nil, err` | 先 `mapGrpcError` 再包装 |
| BFF 区分库存/重复单 | 比 `status.Code(pb.Enum)` | 比 `s.Message()` |
| HTTP 响应 msg | `fromError.Message()` | `fromError.Error()` |
| 新增错误码 | `errcode` 里加一行 `register` | logic 里写 `status.Error(10001)` |
| 共享数字 | contracts proto enum | 两个服务各抄一个 const |

## 给 Agent 的硬规则

Always:

- logic 只返回 `status.Error` / `Wrap` / `ErrorMessage`
- 新增码只在发出方 `internal/errcode` 加一行 `register`
- BFF 默认透传；要分支就比 contracts 里的数字
- 日志可以打 `err` / `Error()`；响应只许 `Message()`

Never:

- 不要 `errors.New` / 裸 `status.Code` 字面量 / 手写 `grpc/status`
- 不要把 `Error()` 写进 HTTP body 或自己拼 gRPC message
- 不要在返回 gRPC 之前再 wrap 一层 `WithStack` / `%w`
- 不要按 message 字符串或裸 `codes.NotFound` 识别「哪种」业务失败
- 不要在 BFF 再抄一份 `codes → http`
- 不要把 `core/status`、文案、`register()` 放进 contracts
