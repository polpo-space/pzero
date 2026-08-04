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
