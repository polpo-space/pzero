# pzero

[![Build Status](https://img.shields.io/github/actions/workflow/status/polpo-space/pzero/ci.yaml?branch=main&label=pzero-ci&logo=github&style=flat-square)](https://github.com/polpo-space/pzero/actions?query=workflow%3Apzero-ci)
[![GitHub release](https://img.shields.io/github/release/polpo-space/pzero.svg?style=flat-square)](https://github.com/polpo-space/pzero/releases/latest)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/polpo-space/pzero)

<p align="center">
<img align="center" width="150px" src="https://oss.jaronnie.com/jzero.svg">
<img align="center" width="300px" src="https://github.com/user-attachments/assets/44184df0-20ce-403d-ab38-74088915bc33">

</p>

**English** | **[简体中文](README.zh-CN.md)**

## Introduction

[pzero](https://github.com/polpo-space/pzero) is a framework developed based on the [go-zero framework](https://github.com/zeromicro/go-zero) and [go-zero/goctl tool](https://github.com/zeromicro/go-zero/tree/master/tools/goctl). It can initialize api/gateway/rpc projects with a single command.

Automatically generate **server and client** framework code based on descriptive files (**api/proto/sql**). With built-in pzero-skills, AI can generate business logic code that follows best practices, reducing development cognitive load and freeing your hands!

Key features:

* Flexible control of pzero configurations through **configuration files/command-line arguments/environment variables**, simple commands to generate code, AI-friendly
* Support generating code based on **git changed files**, support generating code for **specified descriptive files** or **ignoring specified descriptive files**, improving code generation efficiency for large projects
* Built-in common development templates with enhanced template features, support for **custom templates**, building exclusive enterprise internal code templates, greatly reducing development costs

For inherited framework documentation, see the [upstream jzero documentation](https://docs.jzero.io).

## Design Philosophy

* **Developer Experience**: Provide a simple, easy-to-use, production-ready solution to enhance developer experience
* **Template Driven**: All code generation is based on template rendering, default generation follows best practices, and supports custom template content
* **Ecosystem Compatibility**: Does not modify go-zero and go-zero/goctl, maintains ecosystem compatibility, while solving existing pain points and extending new features
* **Team Development**: Team development friendly through module **layering** and **plugin** design
* **Interface Design**: No dependency on specific databases/caches/configuration centers and other infrastructure, choose freely according to actual needs

For inherited framework documentation, see the [upstream jzero documentation](https://docs.jzero.io).

## Quick Start

```shell
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
# Access swagger ui
http://localhost:8001/swagger
```

### docker

```shell
# One-click create project
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest new your_project
cd your_project
# Download dependencies
go mod tidy
# Start server
go run main.go server
# Access swagger ui
http://localhost:8001/swagger
```

For more example code, please visit: https://github.com/jzero-io/examples

## Ecosystem

The inherited upstream ecosystem remains compatible:

* jzero-intellij: https://github.com/jzero-io/jzero-intellij
* jzero-admin: https://github.com/jzero-io/jzero-admin
* templates: https://templates.jzero.io

For more ecosystem, please visit: [https://docs.jzero.io/ecosystem/](https://docs.jzero.io/ecosystem/)

## Contributors

[Contribute](CONTRIBUTING.md)

<a href="https://openomy.app/github/polpo-space/pzero" target="_blank" style="display: block; width: 100%;" align="center">
  <img src="https://openomy.app/svg?repo=polpo-space/pzero&chart=bubble&latestMonth=3" target="_blank" alt="Contribution Leaderboard" style="display: block; width: 100%;" />
</a>

## Stargazers over time

[![Star History Chart](https://api.star-history.com/svg?repos=polpo-space/pzero&type=Date)](https://star-history.com/#polpo-space/pzero&Date)

## Disclaimer

pzero is released under the MIT License and is provided completely free of charge. The authors and contributors assume no liability for any direct or indirect consequences arising from the use of this software, including but not limited to performance degradation, data loss, service interruptions, or any other type of damage.

No Warranty: This software comes with no express or implied warranties, including but not limited to fitness for a particular purpose, non-infringement, merchantability, and reliability.

User Responsibility: By using this software, you understand and agree to assume all risks and responsibilities associated with its use.
