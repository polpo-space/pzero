---
name: pzero-skills
description: Comprehensive knowledge base for pzero framework (enhanced go-zero). Use this skill when working with pzero to understand correct patterns for REST APIs (Handler/Logic/Context architecture), RPC services (service discovery, load balancing), Gateway services, database operations (sqlx and caching), resilience patterns (circuit breaker, rate limiting), and pzero-specific features (git-change-based generation, flexible configuration, custom templates). Essential for generating production-ready pzero code that follows framework conventions.
license: Apache-2.0
---

# pzero Skills for AI Agents

Structured knowledge base optimized for AI agents to help developers work effectively with the [pzero](https://github.com/polpo-space/pzero) framework (enhanced go-zero).

## Overview

This skill provides AI agents with comprehensive pzero knowledge to:
- Generate accurate code following pzero conventions
- Understand the three-layer architecture (Handler -> Logic -> Model)
- Apply best practices for microservices development
- Use pzero-specific features
- Build production-ready applications

## Quick Start

When helping with pzero development:

1. **For new projects**: Start with [Development Workflows](#development-workflows)
2. **For REST APIs**: Check [REST API File Structure](references/rest-api-patterns/api-file-structure.md) - critical rules
3. **For RPC services**: Review [Proto File Structure](references/rpc-patterns/proto-file-structure.md)
4. **For databases**: Review [Database Best Practices](references/database-patterns/best-practices.md)
5. **For SQL changes**: Check [SQL Migration Guide](references/database-patterns/sql-migration.md)
6. **For specific operations**: Reference the appropriate pattern guide below

## Core Patterns

### REST API Development

**Reference**: [references/rest-api-patterns/api-file-structure.md](references/rest-api-patterns/api-file-structure.md)

- API file structure with required settings (`go_package`, `group`, `compact_handler`)
- Three-layer architecture (Handler -> Logic -> Model)
- Request/response type definitions with validation
- Handler patterns and HTTP concerns
- Logic patterns and business implementation
- Correct vs incorrect patterns with examples

### RPC Services

- [Proto File Structure](references/rpc-patterns/proto-file-structure.md): Proto standards, multi-proto support, file structure, HTTP gateway, OpenAPI docs
- [Proto Field Validation](references/rpc-patterns/proto-validation.md): Field validation with protovalidate, CEL expressions, built-in constraints
- [Proto Middleware](references/rpc-patterns/proto-middleware.md): HTTP/RPC middleware at service and method levels

### Database Operations

- [Best Practices](references/database-patterns/best-practices.md): Model import rules, condition chain usage, error handling, field constants
- [SQL Migration Guide](references/database-patterns/sql-migration.md): Managing schema changes with up/down migrations
- [Model Generation](references/database-patterns/model-generation.md): Generate models from SQL files or remote datasource
- [Database Connection](references/database-patterns/database-connection.md): MySQL, PostgreSQL, SQLite, Redis configuration
- [CRUD Operations](references/database-patterns/crud-operations.md): Generated methods and usage patterns

**Critical reminder**: Always use `condition.NewChain()`, never use `condition.New()`.

## Development Workflows

### Creating a New REST API Endpoint

1. Define API specification in an `.api` file with required settings
2. Generate code: `pzero gen --desc desc/api/user.api`
3. Implement logic in `internal/logic/` following the three-layer architecture

### Implementing Database Operations

**Local SQL Mode**:
1. Create or update SQL schema in `desc/sql/*.sql`
2. Create migration files in `desc/sql_migration/`
3. Apply migrations with `pzero migrate up`
4. Generate models with `pzero gen --desc desc/sql/users.sql`

**Remote Datasource Mode**:
1. Create migration files in `desc/sql_migration/`
2. Apply migrations
3. Generate models with `pzero gen`

Common steps:
- Inject model into ServiceContext
- Use the condition builder in logic layer
- Handle errors properly

### Setting Up Database Connection

1. Configure `etc/etc.yaml`
2. Initialize database connection in `ServiceContext` with `modelx.MustNewConn`
3. Register models in the `Model` struct

## Project Structure

```text
skills/
└── pzero-skills/
    ├── SKILL.md
    └── references/
```

Typical pzero project structure:

```text
myproject/
├── skills/
├── .pzero.yaml
├── desc/
│   ├── api/
│   ├── sql/
│   ├── sql_migration/
│   └── proto/
├── internal/
│   ├── handler/
│   ├── logic/
│   ├── model/
│   ├── svc/
│   ├── config/
│   └── middleware/
└── etc/
    └── etc.yaml
```

## Key Principles

### Always Follow

- Always use `condition.NewChain()`
- Keep Handler -> Logic -> Model separation
- Set `go_package`, `group`, and `compact_handler: true` in `.api` files
- Import models with aliases like `xxmodel`
- Use `errors.Is(err, model.ErrNotFound)` from `github.com/pkg/errors`
- Run `pzero gen --desc` before implementing logic
- Use generated field constants

### Never Do

- Never use `condition.New()`
- Never put business logic in handlers
- Never skip `go_package`, `group`, or `compact_handler`
- Never import models without aliases
- Never compare model errors with `==`
- Never hard-code configuration values
- Never implement logic before generating framework code

## Resources

- [Upstream documentation](https://docs.jzero.io)
- [GitHub repository](https://github.com/polpo-space/pzero)
- [Examples](https://github.com/jzero-io/examples)
- [Base framework](https://github.com/zeromicro/go-zero)
