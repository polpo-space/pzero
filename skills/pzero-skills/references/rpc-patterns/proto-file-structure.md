# Proto File Structure

## Overview

pzero supports multi-proto management. By default it scans `desc/proto/`.
For monorepo central contracts, set `gen.proto-dir` to the contract roots.
PB ownership is inferred from each file's `go_package`:
relative (`./types/...`) → generate local pb; absolute (`github.com/.../contracts/gen/...`) → import shared stubs only.

## Standards

- Different modules should live in different proto files
- Service method input and output messages must be defined in the current file
- Set `go_package` explicitly
- Local pb: relative `go_package` like `./types/user`
- Shared pb: absolute `go_package` pointing at shared module, e.g. `github.com/org/repo/contracts/gen/user/v1`

## Example (local)

```protobuf
syntax = "proto3";

package user;

import "google/api/annotations.proto";

option go_package = "./types/user";

message GetUserRequest {
  int64 id = 1;
}

message GetUserResponse {
  int64 id = 1;
  string name = 2;
}

service User {
  rpc GetUser(GetUserRequest) returns(GetUserResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{id}"
    };
  };
}
```

## Central contracts

```yaml
# apps/service/nfc-svc/.pzero.yaml
style: go_zero

gen:
  proto-dir:
    - ../../../contracts/proto/nfc      # absolute go_package → contracts/gen
    - ../../../contracts/proto/user     # relative go_package → local internal/types
```

```bash
# repo root: generate shared pb
make proto

# service: generate server/logic stubs (and local user pb if needed)
cd apps/service/nfc-svc
pzero gen
```

`internal/server/server.go` is always rewritten by gen with `RegisterZrpcServer`.
Optional `gen.proto-include` adds extra `-I` paths; parent of each `proto-dir` is already included.

## Code Generation

```bash
pzero gen --desc desc/proto/user.proto
pzero gen
pzero gen --proto-dir ../../../contracts/proto/nfc --proto-dir ../../../contracts/proto/user
```

## Business Errors

RPC projects ship the same `internal/errcode/errcode.go` as API projects. A `status.Status` implements
`GRPCStatus()`, so it can be returned from any gRPC handler directly: `core/status` maps the business code to a
gRPC code (`400 -> InvalidArgument`, `404 -> NotFound`, `500 -> Internal`, other codes -> `Unknown`) and carries
the original code in a `google.rpc.ErrorInfo` detail. `GRPCStatus()` sends `Message()` only; wrapped causes
stay on the server (`Error()` / logs). Override the default gRPC mapping with `status.WithGRPCCode`.

```go
// internal/errcode/errcode.go
var UserNotFound = register(10001, "user not found")

// internal/logic/user/get_user.go
if errors.Is(err, usermodel.ErrNotFound) {
    return nil, status.Error(errcode.UserNotFound)
}
if err != nil {
    return nil, status.Wrap(errcode.Internal, err)
}
```

On the caller side (an API service calling this RPC), pass the error through unchanged. `status.FromError` in the
response middleware reads the `ErrorInfo` detail and restores `10001`; plain gRPC errors without the detail are
mapped back by gRPC code (`NotFound -> 404`, `PermissionDenied -> 403`, `FailedPrecondition -> 400`,
`Aborted -> 409`, `Canceled -> 499`, otherwise `500`).

```go
resp, err := l.svcCtx.UserRpc.GetUser(l.ctx, &userpb.GetUserRequest{Id: req.Id})
if err != nil {
    return nil, err
}
```

Rules:

- Never build `grpc/status` errors by hand in logic; use `errcode` so HTTP and gRPC share one code space
- Only codes mapped to `Internal`/`Unavailable`/`DeadlineExceeded`/`ResourceExhausted`/`Unimplemented`/`DataLoss`
  count as failures for the go-zero breaker; business errors (including `Unknown` / `FailedPrecondition`) do not trip it
- Codes a BFF must branch on belong in a `contracts` proto enum, not in the service's `internal/errcode` alone;
  see [Shared Error Codes](error-codes.md)
