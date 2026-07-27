# pzero

[![Build Status](https://img.shields.io/github/actions/workflow/status/polpo-space/pzero/ci.yaml?branch=main&label=pzero-ci&logo=github&style=flat-square)](https://github.com/polpo-space/pzero/actions?query=workflow%3Apzero-ci)
[![GitHub release](https://img.shields.io/github/release/polpo-space/pzero.svg?style=flat-square)](https://github.com/polpo-space/pzero/releases/latest)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/polpo-space/pzero)

<p align="center">
<img align="center" width="300px" src="https://github.com/user-attachments/assets/44184df0-20ce-403d-ab38-74088915bc33">

</p>

## 简介

基于 [go-zero框架](https://github.com/zeromicro/go-zero) 以及 [go-zero/goctl工具](https://github.com/zeromicro/go-zero/tree/master/tools/goctl) 开发的 [pzero](https://github.com/polpo-space/pzero) 框架, 可一键初始化 api/gateway/rpc 项目。

基于可描述文件(**api/proto/sql**)自动生成**服务端和客户端**框架代码, 基于内置的 pzero-skills 让 AI 生成符合最佳实践的业务逻辑代码，降低开发心智, 解放双手!

具备以下特点:

* 支持通过**配置文件/命令行参数/环境变量**组合的方式灵活控制 pzero 的各项配置, 极简指令生成代码, ai 友好
* 支持基于 **git 对改动文件**生成代码, 支持对**指定描述文件**生成代码或**忽略指定描述文件**生成代码, 提升大型项目代码生成效率
* 内置常用开发模板并增强模板特性, 支持**自定义模板**, 构建专属企业内部代码模板, 极大降低开发成本

## 设计理念

* **开发体验**: 提供简单好用的一站式生产可用的解决方案, 提升开发体验感
* **模板驱动**: 所有代码生成均基于模板渲染, 默认生成即最佳实践, 且支持自定义模板内容
* **生态兼容**: 不修改 go-zero 和 go-zero/goctl, 保持生态兼容, 同时解决已有的痛点问题并扩展新的功能
* **团队开发**: 通过模块**分层**, **插件**设计, 团队开发友好
* **接口设计**: 不依赖特定数据库/缓存/配置中心等基础设施, 根据实际需求自由选择

## 快速开始

```shell
# 安装 pzero
go install github.com/polpo-space/pzero/cmd/pzero@latest
# 一键安装所需的工具
pzero check
# 一键创建项目
pzero new your_project
cd your_project
# 下载依赖
go mod tidy
# 启动服务端程序
go run main.go server
# 访问 swagger ui
http://localhost:8001/swagger
```

### docker

```shell
# 一键创建项目
docker run --rm -v ${PWD}:/app ghcr.io/polpo-space/pzero:latest new your_project
cd your_project
# 下载依赖
go mod tidy
# 启动服务端程序
go run main.go server
# 访问 swagger ui
http://localhost:8001/swagger
```

## 贡献者

[贡献](CONTRIBUTING.md)

<a href="https://openomy.app/github/polpo-space/pzero" target="_blank" style="display: block; width: 100%;" align="center">
  <img src="https://openomy.app/svg?repo=polpo-space/pzero&chart=bubble&latestMonth=3" target="_blank" alt="Contribution Leaderboard" style="display: block; width: 100%;" />
</a>

## Stargazers over time

[![Star History Chart](https://api.star-history.com/svg?repos=polpo-space/pzero&type=Date)](https://star-history.com/#polpo-space/pzero&Date)

## 免责声明

pzero 基于 MIT License 发布，完全免费提供。作者及贡献者不对使用本软件所产生的任何直接或间接后果承担责任，包括但不限于性能下降、数据丢失、服务中断、或任何其他类型的损害。

无任何保证：本软件不提供任何明示或暗示的保证，包括但不限于对特定用途的适用性、无侵权性、商用性及可靠性的保证。

用户责任：使用本软件即表示您理解并同意承担由此产生的一切风险及责任。
