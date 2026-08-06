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

package version;

import "google/api/annotations.proto";

option go_package = "./types/version";

message VersionRequest {}

message VersionResponse {
  string version = 1;
}

service Version {
  rpc Version(VersionRequest) returns(VersionResponse) {
    option (google.api.http) = {
      get: "/version"
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
    - ../../../contracts/proto/version  # relative go_package → local internal/types
```

```bash
# repo root: generate shared pb
make proto

# service: generate server/logic stubs (and local version pb if needed)
cd apps/service/nfc-svc
pzero gen
```

`internal/server/server.go` is always rewritten by gen with `RegisterZrpcServer`.
Optional `gen.proto-include` adds extra `-I` paths; parent of each `proto-dir` is already included.

## Code Generation

```bash
pzero gen --desc desc/proto/user.proto
pzero gen
pzero gen --proto-dir ../../../contracts/proto/nfc --proto-dir ../../../contracts/proto/version
```
