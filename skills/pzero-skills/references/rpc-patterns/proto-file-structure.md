# Proto File Structure

## Overview

pzero supports multi-proto management. It automatically recognizes files under `desc/proto/` and registers generated services.

## Standards

- Different modules should live in different proto files
- Service method input and output messages must be defined in the current file
- Set `go_package` explicitly

## Example

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

## Code Generation

```bash
pzero gen --desc desc/proto/user.proto
pzero gen
```
