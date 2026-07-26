---
title: 欢迎贡献👏
icon: /icons/ooui-user-contributions-ltr.svg
star: true
order: 30
---

欢迎参与 pzero 的开发以及维护, 这是一件非常有意义的事情, 让我们一起让 pzero 变得更好.

## 步骤

### 1. fork pzero

[点击这里 fork](https://github.com/polpo-space/pzero/fork)

### 2. clone

```shell
git clone https://github.com/your_username/pzero
```

### 3. checkout branch

```shell
cd pzero

git checkout -b feat/patch-1
```

### 4. format the code what you changes

```shell
pzero format
```

### 5. lint codes

```shell
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run --fix
```

### 6. push

```shell
git add .
git commit -m "feat(xx): custom message"
git push
```

### 7. pull request

Create your pull request!!!

## debug pzero

1. fork pzero 并 clone pzero 到本地后

```shell
cd pzero
go install
```

2. new project with frame, e.g. `api`

```shell
pzero new your_project --frame api
```

3. run pzero gen with debug mode

```shell
pzero gen --debug --debug-sleep-time 15
```

4. attach pzero process

推荐采用 goland, 使用 attach 到 pzero 的进程中, 即可 debug, 如下所示:

<video width="720" height="450" controls>
  <source src="https://oss.jaronnie.com/iShot_2024-09-20_09.22.54.mp4" type="video/mp4">
</video>







