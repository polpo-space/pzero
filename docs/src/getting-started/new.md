---
title: Project Initialization
icon: /icons/mdi-new-box.svg
order: 3
---

## Template Introduction

A template is a predefined set of code structures that provides the basic architecture and engineering standards for a project.

Templates help you quickly start initializing a project without writing code from scratch.

## Template Types

pzero provides the following types of templates to meet various scenarios:

* Built-in template(frame): Built-in template providing core framework capabilities, supports optional features (database/cache)
* Path template(home): Specify a path as a template, usually placed inside a specific project to meet specific project needs
* Local template(local): Local global template located in ~/.pzero/templates/local folder
* Remote repository template(remote+branch): Can be used to build enterprise-specific remote template repositories

For detailed usage, see: [Template Guide](../guide/template.md)

## Template market

If built-in frames such as `api`, `rpc`, and `gateway` are not enough, visit the [pzero Template Market](https://templates.jzero.io).

The template market is the entry for discovering built-in templates, official external templates, and third-party templates. You can use it to quickly find the source repository, usage guide, and recommended initialization command for a template.

For official external templates, `pzero new` usually only needs the `--branch` parameter because the default remote repository already points to `https://github.com/jzero-io/templates`.

Quick links:

* [Template market home](https://templates.jzero.io)
* [CLI template](https://templates.jzero.io/external/cli/)
* [Vercel API template branch](https://github.com/jzero-io/templates/tree/api-vercel)

```shell
# Official CLI template
pzero new mycli --branch cli

# Official Vercel API template
pzero new myvercel --branch api-vercel

# Third-party or private template
pzero new your_project --remote <template-repo> --branch <template-branch>
```

For more template details and examples, see the [pzero Template Market](https://templates.jzero.io) and [Template Guide](../guide/template.md).

## Initialize api project

::: code-tabs#shell

@tab pzero cli

```bash
pzero new your_project --frame api
cd your_project
# download dependencies
go mod tidy
# start server
go run main.go server
# visit swagger ui
http://localhost:8001/swagger
```

@tab pzero Docker

```bash
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest new your_project --frame api
cd your_project
# download dependencies
go mod tidy
# start server
go run main.go server
# visit swagger ui
http://localhost:8001/swagger
```
:::

## Initialize rpc project

::: code-tabs#shell

@tab pzero cli

```bash
pzero new your_project --frame rpc
cd your_project
# download dependencies
go mod tidy
# start server
go run main.go server
```

@tab pzero Docker

```bash
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest new your_project --frame rpc
cd your_project
# download dependencies
go mod tidy
# start server
go run main.go server
```
:::

## Optional features model/redis/model+redis

Based on optional features, provides a complete solution for using model/redis/model

```shell
# Use case: need to connect to relational database(model) with database cache(cache), redis
pzero new your_project --features model,cache,redis

# Use case: need to connect to relational database(model), redis
pzero new your_project --features model,redis

# Use case: need to connect to relational database(model) with database cache(cache)
pzero new your_project --features model,cache

# Use case: need to connect to relational database(model)
pzero new your_project --features model

# Use case: merge in-process scheduled jobs with gRPC (ServiceGroup)
pzero new your_project --frame rpc --features job
```
