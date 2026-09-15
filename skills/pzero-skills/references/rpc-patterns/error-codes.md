# Shared Error Codes (Service <-> BFF)

The full API/RPC/BFF playbook is [Error Handling](../error-handling.md). This page is only the contracts proto-enum pattern for numbers a BFF must distinguish.

## Overview

One code space for the whole system. A business error is an integer that travels from the RPC service to the
BFF and on to the client unchanged. `core/status` carries it over gRPC in a `google.rpc.ErrorInfo` detail and
restores it with `status.FromError`, so no layer needs to inspect gRPC codes or message strings.

Each piece of information lives in exactly one place:

| Concern                         | Owner                                        | Form                                   |
| ------------------------------- | -------------------------------------------- | -------------------------------------- |
| Number and meaning              | `contracts/proto/<svc>/v1/error_code.proto`  | proto `enum`, generated into `contracts/gen` |
| Default message and registration| emitting service `internal/errcode`          | one `register(...)` line per code      |
| Number ranges per service       | `contracts` test                             | descriptor assertion                   |
| gRPC / HTTP mapping             | `core/status`                                | built in, never hand-written           |
| Public wording override         | BFF (optional)                               | `status.ErrorMessage` / own `errcode`  |

`contracts` stays dependency-free: it holds numbers only, never `core/status` types, messages, or `register()`.

## Step 1: Declare the enum in contracts

```protobuf
// contracts/proto/order/v1/error_code.proto
syntax = "proto3";

package order.v1;

option go_package = "github.com/org/repo/contracts/gen/order/v1";

// OrderError is the stable business error code space of order-svc.
// Range: 12000-12999. Numbers are never reused or renumbered.
enum OrderError {
  ORDER_ERROR_UNSPECIFIED = 0;
  ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE = 12001;
  ORDER_ERROR_PREPRODUCTION_PENDING_ORDER_EXISTS = 12002;
}
```

Rules:

- One enum per service, named `<Svc>Error`, values prefixed `<SVC>_ERROR_`, value `0` is `UNSPECIFIED`
- Each service owns a fixed thousand-range (`12xxx` order, `13xxx` device, ...); record it in the enum comment
- Add codes only for errors a caller must distinguish; generic `400/401/403/404/500` stay in each project's
  local `internal/errcode` and are not part of the contract
- Never delete or renumber a value; mark it `reserved` if retired

## Step 2: Register message in the emitting service

```go
// apps/service/order-svc/internal/errcode/errcode.go
package errcode

import (
	"net/http"

	"github.com/polpo-space/pzero/core/status"
	"google.golang.org/grpc/codes"

	orderv1 "github.com/org/repo/contracts/gen/order/v1"
)

var (
	InvalidParam = register(http.StatusBadRequest, "invalid parameter")
	NotFound     = register(http.StatusNotFound, "not found")
	Internal     = register(http.StatusInternalServerError, "internal server error")

	PreproductionStockUnavailable   = register(code(orderv1.OrderError_ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE), "preproduction stock unavailable")
	PreproductionPendingOrderExists = register(code(orderv1.OrderError_ORDER_ERROR_PREPRODUCTION_PENDING_ORDER_EXISTS), "a pending preproduction order already exists")
)

func code(e orderv1.OrderError) status.Code { return status.Code(e) }

func register(code status.Code, message string, opts ...status.Option) status.Code {
	status.Register(code, append([]status.Option{status.WithMessage(message)}, opts...)...)
	return code
}
```

The message is the service's business; changing wording never touches `contracts`.

## Step 3: Return it from logic

```go
// apps/service/order-svc/internal/logic/ordercommand/create_order.go
if stock < req.GetQuantity() {
	return nil, status.Error(errcode.PreproductionStockUnavailable)
}
if err != nil {
	return nil, status.Wrap(errcode.Internal, err)
}
```

`status.Status` implements `GRPCStatus()`; return it directly, no `grpc/status` calls in logic.
`Wrap` keeps the cause for server logs (`Error()`); the gRPC message is `Message()` only.

Need a canonical gRPC code other than the default mapping:

```go
StockUnavailable = register(
    code(orderv1.OrderError_ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE),
    "preproduction stock unavailable",
    status.WithGRPCCode(codes.FailedPrecondition),
)
```

## Step 4: Consume in the BFF

Pass through by default; the response middleware restores the original code and message:

```go
resp, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, in)
if err != nil {
	return nil, err
}
```

Branch on a specific code by comparing numbers from the contract. No registration is needed on the BFF side:

```go
import (
	"github.com/polpo-space/pzero/core/status"

	orderv1 "github.com/org/repo/contracts/gen/order/v1"
)

if status.FromError(err).Code() == status.Code(orderv1.OrderError_ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE) {
	return &types.CreateOrderResponse{SoldOut: true}, nil
}
```

Remap only when the BFF deliberately changes the public contract (different code or wording):

```go
if status.FromError(err).Code() == status.Code(orderv1.OrderError_ORDER_ERROR_PREPRODUCTION_STOCK_UNAVAILABLE) {
	return nil, status.ErrorMessage(errcode.SoldOut, "This item is sold out")
}
```

## Step 5: Guard the ranges in contracts

`protoc` only guarantees uniqueness inside one enum. A descriptor test is the range registry:

```go
// contracts/proto/error_code_ranges_test.go
package contracts_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	devicev1 "github.com/org/repo/contracts/gen/device/v1"
	orderv1 "github.com/org/repo/contracts/gen/order/v1"
)

func TestErrorCodeRanges(t *testing.T) {
	ranges := []struct {
		enum     protoreflect.EnumDescriptor
		min, max protoreflect.EnumNumber
	}{
		{orderv1.OrderError(0).Descriptor(), 12000, 12999},
		{devicev1.DeviceError(0).Descriptor(), 13000, 13999},
	}

	for _, r := range ranges {
		values := r.enum.Values()
		for i := 0; i < values.Len(); i++ {
			v := values.Get(i)
			if v.Number() == 0 {
				continue
			}
			require.True(t, v.Number() >= r.min && v.Number() <= r.max,
				"%s = %d outside %d-%d", v.FullName(), v.Number(), r.min, r.max)
		}
	}
}
```

Adding a service means adding one line to `ranges`. Disjoint ranges make cross-enum collisions impossible.

## Wire behaviour

- `Error()` is for logs and the Go error chain and may include the wrapped cause. `Message()` is the
  client-facing text and never includes the cause. `GRPCStatus()` and the HTTP serializer both use `Message()`.
  `status.Wrap(errcode.Internal, err)` therefore logs the DB/SDK error locally and still returns
  `"internal server error"` on the wire
- Contract codes map to gRPC `Unknown` unless registered with `status.WithGRPCCode`. HTTP-semantic local codes
  map by meaning (`400 -> InvalidArgument`, `404 -> NotFound`, `500 -> Internal`, ...). Legacy gRPC errors
  without `ErrorInfo` map back by grpc-gateway rules (`FailedPrecondition`/`OutOfRange -> 400`, `Aborted -> 409`,
  `Canceled -> 499`, `Unknown`/`DataLoss -> 500`)
- go-zero zrpc breaker treats `DeadlineExceeded`/`Internal`/`Unavailable`/`DataLoss`/`Unimplemented`/
  `ResourceExhausted` as failures; other gRPC codes (including `Unknown` and `FailedPrecondition`) do not trip it
- Consumers without `core/status` (other languages, plain gRPC clients) see the gRPC code plus the `ErrorInfo`
  detail with `domain = "pzero"` and `reason = "<number>"`

## Never Do

- Never compare `s.Message()` or `err.Error()` to identify an RPC error
- Never put `fromError.Error()` / `err.Error()` into an HTTP `msg` or gRPC status message; those include causes
- Never branch on bare `codes.NotFound` when the caller needs to know *what* was not found
- Never hand-write `codes -> http` switches in the BFF; `status.FromError` already contains that table
- Never wrap `status.Error` / `status.Wrap` with `errors.WithStack` or `fmt.Errorf("%w")` before returning from a
  gRPC handler: grpc-go then replaces the status message with `err.Error()`, which includes the cause
- Never put `core/status`, `register()`, or messages into `contracts`
- Never re-register contract codes in the BFF unless it deliberately emits them itself
