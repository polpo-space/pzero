# SQL Migration Guide

## Rules

1. Every schema change needs migrations
2. Always provide both `up.sql` and `down.sql`
3. Development can use `pzero migrate`
4. Production should use startup migrations

## Example

```text
desc/sql_migration/
├── 1_create_users_table.up.sql
├── 1_create_users_table.down.sql
├── 2_add_email_index.up.sql
└── 2_add_email_index.down.sql
```

## Guidance

- Never rewrite an old shipped migration
- Keep one logical change per migration
- Test rollbacks
- Maintain consecutive numbering
