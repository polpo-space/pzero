---
title: Error Handling Guide
icon: /icons/eos-icons-api.svg
star: true
order: 0.3
---

# Error Handling Guide

This page is the single source of truth for errors in API, RPC, and BFF logic. Do not invent a second error system.

## One rule

A business error is an integer. The emitting service registers the number and the public message in `internal/errcode`. Logic returns only `status.Error` / `status.Wrap` / `status.ErrorMessage`. HTTP `msg` and the gRPC message are `Message()` only. The wrapped cause stays in process logs.

## Two strings

| Method | Audience | Includes cause? |
| --- | --- | --- |
| `Error()` | Logs, Go error chain | Yes. `Wrap(Internal, dbErr)` becomes `internal server error: pq: ...` |
| `Message()` | Clients, HTTP `msg`, gRPC message | No. A 500 with no explicit text is always `internal server error` |

The generated API middleware already uses `Msg: fromError.Message()`. Do not change it back to `Error()`.

`status.Wrap` returns `Status` itself. Do not wrap it again with `errors.WithStack` or `fmt.Errorf("%w")` before returning from a gRPC handler: grpc-go replaces the status message with `err.Error()` on the wrapped path, and the cause goes on the wire.

## Path

```text
logic
  return status.Error(errcode.X)          // or Wrap / ErrorMessage
        │
        ├─ API: ErrorMiddleware
        │     logs Error() (with cause)
        │     body = { code: X, msg: Message(), data: null }   HTTP 200
        │
        └─ RPC: Status.GRPCStatus()
              gRPC code ← mapping table or WithGRPCCode
              message   ← Message()
              details   ← ErrorInfo{ domain: "pzero", reason: "X" }
                    │
                    └─ BFF: return nil, err          // default: pass through
                          status.FromError(err)      // restores X and Message()
```

The BFF does not need to register the RPC code in its own `errcode` to read the number and message from `FromError`.

## Decision tree

On an `err`, take exactly one branch, in order:

1. **This service's own business failure** (not found, illegal state, semantic bad input) → `return nil, status.Error(errcode.Xxx)`. Override the text for this call → `ErrorMessage`.
2. **Infrastructure failure** (DB, cache, SDK, unknown err) → `return nil, status.Wrap(errcode.Internal, err)`.
3. **Model `ErrNotFound` that is a public 404** → `return nil, status.Error(errcode.NotFound)`.
4. **Error from a downstream RPC**
   - The client should see the downstream code and text → `return nil, err`.
   - This handler must branch on "out of stock" vs "pending order exists" → `status.FromError(err).Code() == status.Code(pb.Xxx)`. Numbers come from a contracts enum, not from message strings.
   - This handler must change the public code or wording → `status.ErrorMessage(errcode.Y, "public text")`.
5. **Never** `errors.New("user not found")`, **never** `grpc/status.Error(codes.NotFound, ...)`, **never** a handwritten `codes → http` switch.

Passing an unregistered code to `status.Error` silently becomes 500. Numbers must come from `register(...)`; do not write a raw `status.Code(10001)`.

## Registering codes

`pzero new --frame api|rpc` generates `internal/errcode/errcode.go`. Declare-and-register:

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

Local generic codes use HTTP semantics (400/401/403/404/500). Service-specific codes that callers must distinguish use a dedicated range, e.g. 10001–10999 for users.

When a gRPC client without `core/status` needs a canonical code, on a service that already depends on this repo's `core/status`:

```go
status.Register(
	12001,
	status.WithMessage("stock unavailable"),
	status.WithGRPCCode(codes.FailedPrecondition),
)
```

The generated `register()` only calls `RegisterWithMessage` so `pzero new` still compiles against the currently published `core/status`.

## Scenario 1: API only (get / register user)

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
		return nil, status.Error(errcode.UserAlreadyExists) // e.g. 10002
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

The client sees:

```json
{ "code": 404, "msg": "not found", "data": null }
```

or `{ "code": 10002, "msg": "user already exists", "data": null }`. HTTP status stays 200; the business status is `code` in the body.

Use `errors.Is` (`github.com/pkg/errors`) for `ErrNotFound`, never `==`.

## Scenario 2: RPC only (same codes, return as gRPC error)

```go
func (l *GetUser) GetUser(in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := l.svcCtx.Model.User.FindOne(l.ctx, in.GetId())
	if errors.Is(err, usermodel.ErrNotFound) {
		return nil, status.Error(errcode.UserNotFound) // e.g. 10001
	}
	if err != nil {
		return nil, status.Wrap(errcode.Internal, err)
	}
	return &userv1.GetUserResponse{Id: user.Id, Name: user.Name}, nil
}
```

Do not write `grpcstatus.Error(codes.NotFound, "user not found")`. `Status` implements `GRPCStatus()`:

- 10001 maps to gRPC `Unknown` by default; the business code is `ErrorInfo.reason`
- 404 maps to `NotFound`, 500 to `Internal`
- the message is `Message()`; SQL text from `Wrap` does not leave the process
- the zrpc server log still prints the full `Error()`

Generated RPC validator middleware already returns `status.ErrorMessage(errcode.InvalidParam, ...)` (`InvalidArgument`).

## Scenario 3: BFF calling RPC

### Pass through (default)

```go
resp, err := l.svcCtx.UserRpc.GetUser(l.ctx, &userv1.GetUserRequest{Id: req.Id})
if err != nil {
	return nil, err
}
```

The client `code` is the RPC's 10001 and `msg` is the RPC's registered text. The BFF does not register 10001.

### Branch on a business code

The number must come from a contracts enum both sides can import. Do not copy it into the BFF `internal/errcode`:

```go
st := status.FromError(err)
if st.Code() == status.Code(orderv1.OrderError_ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE) {
	return &types.CreateOrderResponse{SoldOut: true}, nil
}
return nil, err
```

`errors.Is` cannot match a sentinel from another process. Compare numbers only.

### Remap the public contract

```go
if status.FromError(err).Code() == status.Code(nfcv1.NfcError_NFC_ERROR_TAG_NOT_FOUND) {
	return nil, status.ErrorMessage(errcode.NotFound, "content is not available")
}
```

Remap only to hide internals. Do not run every RPC error through a `mapGrpcError` just to get HTTP 404 — `FromError` already has that table.

### BFF anti-patterns

```go
// Wrong: message string; a wording change breaks the branch
if s.Code() == codes.FailedPrecondition && s.Message() == "preproduction_stock_unavailable" { ... }

// Wrong: too coarse to tell device-missing from outlet-missing
if grpcstatus.Code(err) == codes.NotFound { ... }

// Wrong: copy codes → http in every BFF
switch grpcstatus.Code(err) {
case codes.NotFound:
	return jzerostatus.ErrorMessage(http.StatusNotFound, ...)
}
```

## Shared codes: proto enum in contracts

If a BFF must distinguish an error, the number lives in `contracts/proto/<svc>/v1/error_code.proto`. Not in `internal` (unimportable across modules), not as a string constant in `pkg/`.

```protobuf
syntax = "proto3";
package order.v1;
option go_package = "github.com/org/repo/contracts/gen/order/v1";

// Range: 12000-12999. Never reuse or renumber.
enum OrderError {
  ORDER_ERROR_UNSPECIFIED = 0;
  ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE = 12001;
  ORDER_ERROR_PREPRODUCTION_PENDING_ORDER_EXISTS = 12002;
}
```

Rules:

- `contracts` does not depend on pzero. Numbers only; no messages, no `register()`.
- If `buf.gen.yaml` has `clean: true`, do not put handwritten Go under `contracts/gen`.
- Messages and registration stay in the emitting service `internal/errcode`.
- One thousand-range per service. `protoc` only uniqueness-checks one enum; a contracts test guards ranges.
- Generic 400/401/403/404/500 stay local.
- Retire with `reserved`; never delete or change a number.

Step-by-step (enum, register, BFF compare, range test): [Shared Error Codes](https://github.com/polpo-space/pzero/blob/main/skills/pzero-skills/references/rpc-patterns/error-codes.md).

## Mapping tables (do not copy into business code)

HTTP-semantic code → gRPC:

| Business code | gRPC |
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
| other business ranges | Unknown (override with `WithGRPCCode`) |

Legacy gRPC without `ErrorInfo` → HTTP (grpc-gateway rules):

| gRPC | Business code |
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
| Unknown / Internal / DataLoss / anything else | 500 |

The go-zero zrpc breaker counts `DeadlineExceeded` / `Internal` / `Unavailable` / `DataLoss` / `Unimplemented` / `ResourceExhausted` as failures. Default `Unknown` business codes and `FailedPrecondition` do not trip it.

## Right vs wrong

| Case | Right | Wrong |
| --- | --- | --- |
| User missing | `status.Error(errcode.UserNotFound)` | `errors.New("user not found")` |
| DB down | `status.Wrap(errcode.Internal, err)` | `return nil, err` leaking `pq:` |
| RPC missing | `status.Error(errcode.X)` | `grpcstatus.Error(codes.NotFound, ...)` |
| BFF → RPC | `return nil, err` | `mapGrpcError` then wrap |
| Distinguish stock vs duplicate | compare `status.Code(pb.Enum)` | compare `s.Message()` |
| HTTP `msg` | `fromError.Message()` | `fromError.Error()` |
| New code | one `register` line in `errcode` | `status.Error(10001)` in logic |
| Shared number | contracts proto enum | duplicated consts |

## Hard rules for agents

Always:

- Logic returns only `status.Error` / `Wrap` / `ErrorMessage`
- Add codes as one `register` line in the emitting service `internal/errcode`
- BFF passes through by default; branch on contracts numbers
- Logs may print `err` / `Error()`; responses may only use `Message()`

Never:

- `errors.New`, raw `status.Code` literals, or handwritten `grpc/status`
- `Error()` in an HTTP body or a handmade gRPC message
- `WithStack` / `%w` around a status before returning from gRPC
- Identify a *kind* of business failure by message string or bare `codes.NotFound`
- Copy a `codes → http` switch into the BFF
- Put `core/status`, messages, or `register()` into `contracts`
