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

## Generate models only

`pzero gen model` runs the model stage only. It does not regenerate api or rpc, even when `desc/api` or proto files exist.

The command implies datasource mode. Pass `--model-datasource-url` or set it in `.pzero.yaml`.

```shell
pzero gen model
pzero gen model --model-datasource-url "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"
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

More usage: [pzero guide](../guide/jzero.md)

## Migrating retired generator features

- Remove `--git-change`, `gen.git-change`, and `PZERO_GEN_GIT_CHANGE`. `pzero gen` regenerates the application's descriptors; select applications in CI and keep `--desc` / `--desc-ignore` for explicit scopes. `pzero format --git-change` remains supported.
- Remove `--route2code`, `gen.route2code`, `gen.swagger.route2code`, and the corresponding environment variables. Pzero no longer generates permission-code maps or injects them into Swagger descriptions. Move any authorization mapping to application-owned code, update its callers, then remove the old generated `internal/handler/route2code.go`.
- Generation no longer writes IDE navigation metadata. GoLand integrations that consume `~/.pzero/desc-metadata` must use another navigation source. Existing metadata is left untouched and may be deleted manually.
- Model generation accepts exactly one PostgreSQL URL (a scalar or a one-element YAML list). Multiple URLs and `database.table` names are rejected before model files are written. Use unqualified table names and `model-schema` for the PostgreSQL schema. Generate separate databases in separate projects/output directories; running them successively in one project would overwrite `model.go` registration.
- Custom `model/model.go.tpl` templates must use `.Imports` (strings) and `.TableInfos` for the single database. `.ImportsWithAlias`, `.MutiModels`, `.MutiModelsWithAlias`, and table `Alias`/`FullName`/`WithCache` fields are no longer supplied. Cache expiry settings and individual model constructors remain supported.

Retired generator flags fail as unknown flags; retired configuration/environment keys produce a migration error instead of silently changing behavior.
