---
title: Generate server code
icon: /icons/vscode-icons-folder-type-api-opened.svg
order: 4
---

pzero code generation command is extremely simple, only need `pzero gen` to automatically recognize all descriptor files/configurations and complete code generation.

After adding descriptor files with the `pzero add` command from the previous document, execute `pzero gen` to see the generated files.

## Generate code

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

## Generate code based on git changes

::: tip Get new/modified descriptor files based on git status -su
:::

```shell
pzero gen --git-change
```

## Generate code for specific desc

```shell
pzero gen --desc desc/api/xx.api
pzero gen --desc desc/proto/xx.proto
pzero gen --desc desc/sql/xx.sql
```

## Ignore specific desc when generating code

```shell
pzero gen --desc-ignore desc/api/xx.api
pzero gen --desc-ignore desc/proto/xx.proto
pzero gen --desc-ignore desc/sql/xx.sql
```

For more usage, see: [pzero guide](../guide/pzero.md)
