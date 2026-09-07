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

Model 生成需要在配置中开启 `model-datasource: true` 并提供 `model-datasource-url`。`desc/sql` 仅为 schema snapshot，不会触发 model 生成。

## 基于 git 变动生成代码

::: tip 基于 git status -su 获取新增/改动的可描述文件
:::

```shell
pzero gen --git-change
```

## 指定 desc 生成代码

`--desc` 用于限定 **api/proto** 生成范围；指定后会跳过 model 生成。

```shell
pzero gen --desc desc/api/xx.api
pzero gen --desc desc/proto/xx.proto
```

## 忽略指定 desc 生成代码

```shell
pzero gen --desc-ignore desc/api/xx.api
pzero gen --desc-ignore desc/proto/xx.proto
```

更多用法请参阅: [pzero 指南](../guide/pzero.md)
