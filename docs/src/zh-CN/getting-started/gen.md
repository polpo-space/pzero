---
title: 生成服务端代码
icon: /icons/vscode-icons-folder-type-api-opened.svg
order: 4
---

pzero 生成代码命令极其精简, 仅需 `pzero gen` 就能自动识别所有的可描述文件/配置, 完成代码的生成.

通过上一篇文档的 `pzero add` 命令添加可描述文件后, 执行 `pzero gen` 即可看到生成的文件了.

## 生成代码

::: code-tabs#shell

@tab pzero

```bash
cd your_project
pzero gen
```

@tab Docker

```bash
cd your_project
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest gen
```
:::

## 基于 git 变动生成代码

::: tip 基于 git status -su 获取新增/改动的可描述文件
:::

```shell
pzero gen --git-change
```

## 指定 desc 生成代码

```shell
pzero gen --desc desc/api/xx.api
pzero gen --desc desc/proto/xx.proto
pzero gen --desc desc/sql/xx.sql
```

## 忽略指定 desc 生成代码

```shell
pzero gen --desc-ignore desc/api/xx.api
pzero gen --desc-ignore desc/proto/xx.proto
pzero gen --desc-ignore desc/sql/xx.sql
```

更多用法请参阅: [pzero 指南](../guide/pzero.md)