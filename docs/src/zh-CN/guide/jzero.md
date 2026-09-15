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
style: go_zero
```
:::

即 `pzero gen` + `.pzero.yaml` = `pzero gen --style go_zero`

对于环境变量的使用, 需要增加前缀 `PZERO_`, 如 `PZERO_STYLE`

即 `PZERO_STYLE=go_zero pzero gen` = `pzero gen --style go_zero`

环境变量的定义支持使用配置文件, 默认为 `.pzero.env.yaml`

如:

::: code-tabs#yaml
@tab .pzero.env.yaml
```yaml
PZERO_STYLE: go_zero
```
:::

### 子命令

对于子命令的配置, 如: `pzero gen swagger --output desc/swagger` 对应 `.pzero.yaml` 内容

::: code-tabs#yaml
@tab .pzero.yaml
```yaml
gen:
  swagger:
    output: desc/swagger
```
:::

`pzero gen swagger` + `.pzero.yaml` = `pzero gen swagger --output desc/swagger`

同样支持环境变量的配置

::: code-tabs#yaml
@tab .pzero.env.yaml
```yaml
PZERO_GEN_SWAGGER_OUTPUT: desc/swagger
```
:::

`pzero gen swagger` + `.pzero.env.yaml` = `pzero gen swagger --output desc/swagger`

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

