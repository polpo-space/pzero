# {{ .APP }}

## Install Pzero Framework

```shell
go install github.com/polpo-space/pzero/cmd/pzero@latest

pzero check
```

## Generate code

### Generate server code

```shell
pzero gen
```

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
