---
title: 安装 pzero
icon: /icons/marketeq-download-alt-4.svg
order: 2
---

## 安装 golang

推荐采用 [gvm](https://github.com/jaronnie/gvm) 安装 golang 环境

## 安装 pzero

提供以下三种方式使用 pzero, 请根据实际情况任选一种即可

* 源码安装(**go version >= go1.25.0**)
* 直接[下载 pzero 二进制文件](https://github.com/polpo-space/pzero/releases)
* 基于 Docker 安装 pzero, [镜像地址](https://github.com/polpo-space/pzero/pkgs/container/pzero)

### 源码安装 pzero

```bash
# 设置国内代理(可选)
# go env -w GOPROXY=https://goproxy.cn,direct
go install github.com/polpo-space/pzero/cmd/pzero@latest

# 获取 pzero 版本信息
pzero version

# 自动下载所依赖的工具
pzero check
```

### 下载 pzero 二进制文件

[下载地址](https://github.com/polpo-space/pzero/releases)

根据自己的操作系统选择对应的压缩包, 解压后放在 `$GOPATH/bin` 目录下即可

执行以下内容完成 pzero 的环境准备

```shell
# 获取 pzero 版本信息
pzero version

# 自动下载所依赖的工具
pzero check
```

### 基于 Docker 安装 pzero

```shell
# 获取 pzero 版本信息
docker run --rm ghcr.io/polpo-space/pzero:latest version
```

## 升级 pzero

```shell
# 升级为最新版
pzero upgrade
# 升级到指定版本 
pzero upgrade --channel <commit_hash> 或 <tag>
```