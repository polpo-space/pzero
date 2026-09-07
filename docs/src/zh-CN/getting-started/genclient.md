---
title: 生成客户端代码
icon: /icons/clarity-thin-client-line.svg
order: 5
---

## 生成 Swagger

### 使用方法

::: code-tabs#shell

@tab pzero cli

```bash
cd your_project
pzero gen swagger
# 合并成一个 swagger.json 文件
pzero gen swagger --merge
```

@tab pzero Docker
```bash
cd your_project
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest gen swagger
# 合并成一个 swagger.json 文件
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest gen swagger --merge
```
:::

**Swagger 文件**: `desc/swagger`

**Swagger UI 地址**: `localhost:8001/swagger`

## Zrpc 客户端

`pzero gen zrpcclient` 已移除。

跨服务 RPC 调用请在业务仓自行维护：用 protoc/Buf 生成共享 `*.pb.go` / `*_grpc.pb.go`，再用薄 `Clientset` 把 `zrpc.Client` 装配成各 Service 的 gRPC client（不要再生成本地 `typed/` / `types/` 副本）。
