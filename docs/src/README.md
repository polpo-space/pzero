---
home: false
icon: /icons/fluent-home-heart-20-filled.svg
title: Home
---

<div style="text-align: center;">
  <img src="https://oss.jaronnie.com/jzero.svg" style="width: 33%;" alt=""/>
</div>

## Introduction

**pzero** is a framework developed based on the [go-zero framework](https://github.com/zeromicro/go-zero) and [go-zero/goctl tools](https://github.com/zeromicro/go-zero/tree/master/tools/goctl). It enables one-click initialization of api/gateway/rpc projects.

Based on describable files (**api/proto/sql**), it automatically generates **server-side and client-side** framework code. Combined with built-in pzero-skills, it empowers AI to generate business logic code that follows best practices, reducing development cognitive load and freeing your hands!

Key features include:

* Supports flexible control of pzero configurations through combinations of **configuration files/command-line parameters/environment variables**, with minimal commands to generate code, AI-friendly
* Supports generating code based on **git-changed files** or specific descriptor files, or ignoring specific descriptor files, improving code generation efficiency for large projects
* Built-in common development templates with enhanced template features, supports **custom templates** to build proprietary enterprise code templates, significantly reducing development costs

For more details, please visit: [https://docs.jzero.io](https://docs.jzero.io)

## Design Philosophy

* **Developer Experience**: Provides a simple, easy-to-use, one-stop production-ready solution that enhances the development experience
* **Template-Driven**: All code generation is based on template rendering, with default generation following best practices, and supports custom template content
* **Ecosystem Compatibility**: Does not modify go-zero and go-zero/goctl, maintains ecosystem compatibility while addressing existing pain points and extending new features
* **Team Development**: Through clear module **layering**, it's friendly to team development
* **Interface Design**: Does not depend on specific databases/cache/config centers and other infrastructure, allowing free choice based on actual needs

## Quick Start

::: code-tabs#shell
@tab pzero cli

```bash
# Install pzero
go install github.com/polpo-space/pzero/cmd/pzero@latest
# One-click install required tools
pzero check
# One-click create project
pzero new your_project
cd your_project
# Download dependencies
go mod tidy
# Start server
go run main.go server
# Visit swagger ui
http://localhost:8001/swagger
```

@tab pzero Docker

```bash
# One-click create project
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest new your_project
cd your_project
# Download dependencies
go mod tidy
# Start server
go run main.go server
# Visit swagger ui
http://localhost:8001/swagger
```
:::
