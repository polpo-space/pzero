# Model Generation

## Overview

pzero generates models from a PostgreSQL remote datasource only.

`desc/sql` is a schema snapshot for review/alignment. It may coexist with datasource mode and is never model gen input.

## Datasource Mode

```yaml
gen:
  model-driver: pgx
  model-datasource: true
  model-datasource-url:
    - "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"
  model-datasource-table:
    - users
```

```bash
# migrate first, then generate models from the live schema
pzero gen
# regenerate models only; does not rewrite api/rpc
pzero gen model
```

`--desc` scopes api/proto only and skips model generation on the full `pzero gen` pipeline. Do not pass `.sql` to `--desc`. Use `pzero gen model` when the service already has api/proto and you only need table models.

## PostgreSQL Arrays and Type Overrides

Generate array-bearing tables normally; an array column does not require a handwritten model or a custom registration template. The built-in templates import `github.com/lib/pq` when required.

| PostgreSQL column | Default Go type |
| --- | --- |
| `TEXT[]`, `VARCHAR[]` | `pq.StringArray` |
| `BOOLEAN[]` | `pq.BoolArray` |
| `SMALLINT[]`, `INTEGER[]`, `BIGINT[]` | `pq.Int64Array` |
| `REAL[]`, `DOUBLE PRECISION[]` | `pq.Float64Array` |

Preserve `nil` (SQL NULL) versus a non-nil empty array (`{}`). Default mappings cover one-dimensional arrays with non-NULL elements. For map-based partial updates, pass a `driver.Valuer` such as `pq.StringArray`, rather than a plain `[]string`.

Override types in the application's `goctl.yaml`, using PostgreSQL's introspected type name (`_text` for `TEXT[]`):

```yaml
model:
  types_map:
    _text:
      type: types.StringList
      null_type: types.StringList
      pkg: example.com/app/internal/types
```

The mapped type must support database scanning and writing (`sql.Scanner` and `driver.Valuer`); a `pq.StringArray` alias is one option. Supply the imported package and required dependencies in the application. Prefer this override to replacing an entire model for one field.

## Generated Files and Upgrades

- `*_model_gen.go` is overwritten on regeneration; `*_model.go` preserves handwritten extensions. Generated methods live on the embedded `default...Model`; existing `custom...Model` methods may override them when their signatures match the interface.
- Upgrade the CLI and application runtime together, then regenerate all models. Update project-owned `.template/go-zero/model` templates as a complete set: interface, implementation, imports, and constructor templates depend on each other.
- Migrate removed dialect APIs: `MustNewConnAndSqlbuilderFlavor` to `MustNewConn`; remove `ModelOpts.Flavor` and `modelx.WithFlavor`; replace condition `*WithFlavor` calls with their PostgreSQL equivalents (`BuildSelect`, `BuildUpdate`, `BuildDelete`, or the helper without the `WithFlavor` suffix and flavor argument).
- Apply the explicit driver configuration in [Database Connection](database-connection.md) before upgrading a running application.
- Generated constructors clone fixed PostgreSQL field metadata without repeating reflection or mutating package globals. Handwritten extensions should use the instance metadata and treat it as read-only.
- Each goctl subprocess uses a private workspace. Generation must preserve the application's `go.work`; do not repair a failing test by rewriting the workspace after generation.

See [CRUD Operations](crud-operations.md) for the generated method contracts, especially affected rows and row locks.

## Verification

After a model or template change, generate from a disposable, migrated PostgreSQL database and compile the result. Exercise array NULL/empty/value round trips, conditional-update conflicts, transaction rollback, and lock contention as applicable. When changing shared templates, cover both cached and uncached generation, type overrides, and regeneration preserving handwritten extensions. SQLite does not verify PostgreSQL arrays or locking behavior.

For framework generator changes, run the existing `TestModelSQLImport` matrix and `TestPostgresGeneratedModels` contract in `cmd/pzero/internal/command/gen/genmodel`. Set `GOCTL` to the pinned binary built from the repository and `PZERO_TEST_POSTGRES_DSN` to a disposable database; the database contract skips if either is absent, so a skipped run is not validation. Its generated-model tests use `-race` and verify workspace preservation and checkout module resolution.

## Guidance

- Regenerate models after applying migrations
- Keep `model-datasource-table` complete: `internal/model/model.go` is fully rewritten from that list
- Pair generation with migration discipline (`desc/sql_migration/`)
- Keep `desc/sql` snapshots in sync for review, not for gen
