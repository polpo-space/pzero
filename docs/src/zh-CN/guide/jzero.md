---
title: 玩转 pzero
icon: /icons/catppuccin-astro-config.svg
star: true
order: 0.1
---

## 关于配置

* 支持通过配置文件 .pzero.yaml 控制各种参数
* 支持通过 flag 控制各种参数
* 支持通过环境变量控制各种参数
* 支持通过以上组合的方式控制各种参数, 优先级从高到低为: flag > 环境变量 > 配置文件

如: `pzero gen --style go_zero` 对应 `.pzero.yaml` 内容

::: code-tabs#yaml
@tab .pzero.yaml
```yaml
gen:
  git-change: true
```
:::

即 `pzero gen` + `.pzero.yaml` = `pzero gen --git-change=true`

对于环境变量的使用, 需要增加前缀 `PZERO_`, 如 `PZERO_GEN_GIT_CHANGE`

即 `PZERO_GEN_GIT_CHANGE=go_zero pzero gen` = `pzero gen --git-change=true`

环境变量的定义支持使用配置文件, 默认为 `.pzero.env.yaml`

如:

::: code-tabs#yaml
@tab .pzero.env.yaml
```yaml
PZERO_GEN_GIT_CHANGE: true
```
:::

### 子命令

对于子命令的配置, 如: `pzero gen zrpcclient --output client` 对应 `.pzero.yaml` 内容

::: code-tabs#yaml
@tab .pzero.yaml
```yaml
gen:
  zrpcclient:
    output: client
```
:::

`pzero gen zrpcclient` + `.pzero.yaml` = `pzero gen zrpcclient --output client`

同样支持环境变量的配置 `PZERO_GEN_ZRPCCLIENT_NAME`

::: code-tabs#yaml
@tab .pzero.env.yaml
```yaml
PZERO_GEN_ZRPCCLIENT_OUTPUT: client
```
:::

`pzero gen zrpcclient` + `.pzero.env.yaml` = `pzero gen zrpcclient --output client`

## 设置工作目录

```shell
pzero gen -w /path/to
```

## 设置 quiet 模式

```shell
pzero gen --quiet
```

## 设置 debug 模式

```shell
pzero gen --debug
```

