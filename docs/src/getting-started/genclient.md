---
title: Generate client code
icon: /icons/clarity-thin-client-line.svg
order: 5
---

## Generate Swagger

### Usage

::: code-tabs#shell

@tab pzero cli

```bash
cd your_project
pzero gen swagger
# merge into single swagger.json file
pzero gen swagger --merge
```

@tab pzero Docker
```bash
cd your_project
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest gen swagger
# merge into single swagger.json file
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest gen swagger --merge
```
:::

**Swagger files**: `desc/swagger`

**Swagger UI address**: `localhost:8001/swagger`

## Zrpc client

`pzero gen zrpcclient` has been removed.

For cross-service RPC calls, generate shared `*.pb.go` / `*_grpc.pb.go` with protoc/Buf in your business repo, then maintain a thin `Clientset` that wires `zrpc.Client` to each service's gRPC client. Do not regenerate local `typed/` / `types/` copies.
