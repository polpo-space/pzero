# {{ .APP }}

## Install Pzero Framework

```shell
go install github.com/polpo-space/pzero/cmd/pzero@latest

pzero check
```

## Build with version info

```shell
go build -ldflags "-X '{{.Module}}/internal/buildinfo.Version=v0.1.0' \
  -X '{{.Module}}/internal/buildinfo.Commit=$(git rev-parse --short HEAD)' \
  -X '{{.Module}}/internal/buildinfo.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" .
```

CLI `version`、启动时版本输出、以及 `/api/version` 均读取 `internal/buildinfo` 包，无需再设置环境变量。

## Generate code

### Generate server code

```shell
pzero gen
```

### Generate swagger code

```shell
pzero gen swagger
```

Generated Swagger JSON is written to `desc/swagger`. It is not served by the
API by default; register Swagger routes explicitly if runtime documentation
endpoints are required.

{{ if has "model" .Features }}## Database migrations

Service migrations support PostgreSQL through the `pgx` driver only.

Create a migration file pair:

```shell
go run . migrate create add_example
```

Apply pending migrations explicitly before starting the server:

```shell
go run . migrate up --config etc/etc.yaml
```

Run `go run . migrate --help` for rollback, version, goto, and force commands.

{{ end }}
