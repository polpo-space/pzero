# Proto Middleware

## Overview

pzero supports HTTP and RPC middleware at both service level and method level.

## Examples

Service-level HTTP middleware:

```protobuf
option (jzero.api.http_group) = {
  middleware: "auth,logging",
};
```

Method-level HTTP middleware:

```protobuf
option (jzero.api.http) = {
  middleware: "adminCheck",
};
```

Service-level RPC middleware:

```protobuf
option (jzero.api.zrpc_group) = {
  middleware: "trace",
};
```

## Guidance

- Put shared middleware at service level
- Add only truly specific middleware at method level
- Middleware executes left to right
