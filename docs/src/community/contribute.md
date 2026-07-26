---
title: Welcome to Contribute👏
icon: /icons/ooui-user-contributions-ltr.svg
star: true
order: 30
---

Welcome to participate in pzero's development and maintenance. This is a very meaningful thing. Let's make pzero better together.

## Steps

### 1. fork pzero

[Click here to fork](https://github.com/polpo-space/pzero/fork)

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

1. After forking pzero and cloning pzero locally

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

It's recommended to use goland, attach to pzero's process for debugging, as shown below:

<video width="720" height="450" controls>
  <source src="https://oss.jaronnie.com/iShot_2024-09-20_09.22.54.mp4" type="video/mp4">
</video>
