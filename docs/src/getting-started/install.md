---
title: Install pzero
icon: /icons/marketeq-download-alt-4.svg
order: 2
---

## Install golang

Recommend using [gvm](https://github.com/jaronnie/gvm) to install golang environment

## Install pzero

Provides the following three ways to use pzero, choose one based on your actual situation

* Source installation(**go version >= go1.25.0**)
* Directly [download pzero binary](https://github.com/polpo-space/pzero/releases)
* Install pzero based on Docker, [image address](https://github.com/polpo-space/pzero/pkgs/container/pzero)

### Install pzero from source

```bash
# Set domestic proxy (optional)
# go env -w GOPROXY=https://goproxy.cn,direct
go install github.com/polpo-space/pzero/cmd/pzero@latest

# Get pzero version
pzero version

# Auto download required tools
pzero check
```

### Download pzero binary

[Download address](https://github.com/polpo-space/pzero/releases)

Select the corresponding package based on your operating system, extract and place in `$GOPATH/bin` directory

Execute the following to complete pzero environment setup

```shell
# Get pzero version
pzero version

# Auto download required tools
pzero check
```

### Install pzero based on Docker

```shell
# Get pzero version
docker run --rm ghcr.io/polpo-space/pzero:latest version
```

## Upgrade pzero

```shell
# Upgrade to latest version
pzero upgrade
# Upgrade to specific version
pzero upgrade --channel <commit_hash> or <tag>
```
