# Error Handling (API, RPC, BFF)

Single playbook. Full narrative: docs `guide/error-handling.md` (EN / zh-CN).
Contracts enum details: [error-codes.md](rpc-patterns/error-codes.md).

When writing or reviewing any `return ..., err` in logic, follow this page. Do not invent a second error system.

## Decision tree

Take exactly one branch:

1. This service's own business failure → `return nil, status.Error(errcode.X)` (or `ErrorMessage` to override text).
2. DB / cache / SDK / unknown → `return nil, status.Wrap(errcode.Internal, err)`.
3. `errors.Is(err, xxmodel.ErrNotFound)` that is a public 404 → `status.Error(errcode.NotFound)`.
4. Downstream RPC error:
   - Client should see the downstream code → `return nil, err`.
   - This handler must distinguish two business outcomes → `status.FromError(err).Code() == status.Code(pb.EnumValue)`.
   - This handler must change public code/text → `status.ErrorMessage(errcode.Y, "public text")`.
5. Stop. Never `errors.New`, never `grpc/status.Error(codes.X, ...)`, never a `codes → http` switch.

Unregistered codes silently become 500. Numbers come from `register(...)`, never a raw literal.

## Two strings

- `Error()` — logs only, may include cause (`internal server error: pq: ...`).
- `Message()` — HTTP `msg` and gRPC message. Never includes cause. Bare 500 → `"internal server error"`.

Generated API middleware must stay `Msg: fromError.Message()`.
Do not wrap `status.Error`/`Wrap` with `errors.WithStack` or `fmt.Errorf("%w")` before returning from gRPC: grpc-go then sets the message to `err.Error()`.

## API logic

```go
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
```

HTTP stays 200; body is `{code, msg, data}`. Add new codes as one line in `internal/errcode`.

## RPC logic

Same as API. Return `status.Error` / `Wrap` directly. Do not build `grpc/status`.

- HTTP-semantic codes map (`404 → NotFound`, `500 → Internal`, `400 → InvalidArgument`).
- Other business ranges map to `Unknown`; override with `status.WithGRPCCode` after this `core/status` is the module the service actually uses.
- Cause from `Wrap` does not leave the process; zrpc logs still see `Error()`.

## BFF calling RPC

```go
resp, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, in)
if err != nil {
	return nil, err // default: pass through
}
```

Branch (numbers from contracts, no BFF registration):

```go
if status.FromError(err).Code() == status.Code(orderv1.OrderError_ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE) {
	return &types.CreateOrderResponse{SoldOut: true}, nil
}
```

Remap only to hide internals:

```go
return nil, status.ErrorMessage(errcode.NotFound, "content is not available")
```

Wrong: `s.Message() == "..."`, `grpcstatus.Code(err) == codes.NotFound` when more than one NotFound exists, per-BFF `mapGrpcError`.

## Where the number lives

| Kind | Where |
| --- | --- |
| Generic 400/401/403/404/500 | each project's `internal/errcode` |
| Code a BFF must distinguish | `contracts/proto/<svc>/v1/error_code.proto` enum; message+`register` in the emitting service only |
| Range uniqueness across services | contracts descriptor test |

`contracts` stays pzero-free (numbers only). Do not put handwritten Go in `contracts/gen` if buf `clean: true`.

## Never

- `errors.New` / raw `status.Code(10001)` / handwritten `grpc/status`
- `fromError.Error()` in HTTP `msg`
- wrap status before returning from a gRPC handler
- identify business kind by message string
- copy `codes → http` into a BFF
- put `core/status`, messages, or `register()` into contracts
