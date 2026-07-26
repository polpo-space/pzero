---
title: Mastering pzero
icon: /icons/catppuccin-astro-config.svg
star: true
order: 0.1
---

## About Configuration

* Supports controlling various parameters through configuration file .pzero.yaml
* Supports controlling various parameters through flag
* Supports controlling various parameters through environment variables
* Supports controlling various parameters through combination of above methods, priority from high to low: flag > environment variables > configuration file

Example: `pzero gen --style go_zero` corresponds to `.pzero.yaml` content

::: code-tabs#yaml
@tab .pzero.yaml
```yaml
gen:
  git-change: true
```
:::

`pzero gen` + `.pzero.yaml` = `pzero gen --git-change=true`

For environment variable usage, need to add prefix `PZERO_`, such as `PZERO_GEN_GIT_CHANGE`

`PZERO_GEN_GIT_CHANGE=go_zero pzero gen` = `pzero gen --git-change=true`

Environment variable definition supports using configuration file, default is `.pzero.env.yaml`

Example:

::: code-tabs#yaml
@tab .pzero.env.yaml
```yaml
PZERO_GEN_GIT_CHANGE: true
```
:::

### Subcommands

For subcommand configuration, such as: `pzero gen zrpcclient --output client` corresponds to `.pzero.yaml` content

::: code-tabs#yaml
@tab .pzero.yaml
```yaml
gen:
  zrpcclient:
    output: client
```
:::

`pzero gen zrpcclient` + `.pzero.yaml` = `pzero gen zrpcclient --output client`

Also supports environment variable configuration `PZERO_GEN_ZRPCCLIENT_NAME`

::: code-tabs#yaml
@tab .pzero.env.yaml
```yaml
PZERO_GEN_ZRPCCLIENT_OUTPUT: client
```
:::

`pzero gen zrpcclient` + `.pzero.env.yaml` = `pzero gen zrpcclient --output client`

## Set working directory

```shell
pzero gen -w /path/to
```

## Set quiet mode

```shell
pzero gen --quiet
```

## Set debug mode

```shell
pzero gen --debug
```

## Custom CLI plugins

If built-in commands are not enough, `pzero` can dispatch unknown commands to external executables in `PATH`.

See [Custom CLI plugins](./cli-plugin.md).
