---
title: Generate server code
icon: /icons/vscode-icons-folder-type-api-opened.svg
order: 4
---

pzero keeps code generation minimal: `pzero gen` discovers descriptors and config, then generates code.

After adding descriptors with `pzero add`, run `pzero gen` to see generated files.

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

Model generation requires `model-datasource: true` and `model-datasource-url`. `desc/sql` is only a schema snapshot and does not trigger model generation.

## Generate from git changes

::: tip Uses git status -su for added/changed descriptors
:::

```shell
pzero gen --git-change
```

## Generate with explicit desc

`--desc` scopes **api/proto** generation; model generation is skipped when desc is set.

```shell
pzero gen --desc desc/api/xx.api
pzero gen --desc desc/proto/xx.proto
```

## Ignore descriptors

```shell
pzero gen --desc-ignore desc/api/xx.api
pzero gen --desc-ignore desc/proto/xx.proto
```

More usage: [pzero guide](../guide/pzero.md)
