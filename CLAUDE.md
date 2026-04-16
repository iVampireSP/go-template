## Project Overview

Go project template — a minimal API project with User model, JWT auth, and admin panel.

## Common Commands

### Build & Run

```bash
go build ./...                # Build
go generate ./...             # Run code generation (ent ORM + autodi wiring)
go run . serve api            # Start User API server
go run . serve admin          # Start Admin API server
```

### Database Migrations

```bash
go run . migrate up           # Apply pending migrations
go run . migrate down         # Rollback last migration
go run . migrate fresh        # Drop all + re-run (destructive)
go run . migrate status       # Check migration status
```

### Code Generation

After changing ent schemas or adding/removing `New*` constructors:
```bash
go generate ./ent/...         # Regenerate ent ORM code
go generate ./generate.go     # Regenerate autodi DI wiring (main.go)
```

## Architecture

### Autodi Code Generation

`main.go` is **auto-generated** by `autodi` — do NOT edit it directly. Directives live in `generate.go`. After adding/changing any constructor signature, run `go generate ./generate.go`.

### ORM (ent)

- Schemas in `ent/schema/` — User entity with TimeMixin + SoftDeleteMixin
- Generated code in `ent/` — CRUD, queries, predicates (do NOT edit)
- Migrations in `database/migrations/` — goose SQL migrations

### API Structure

- `internal/api/user/v1/` — User API (auth + profile endpoints)
- `internal/api/admin/v1/` — Admin API (login + user management)
- `internal/api/wellknown/` — JWKS + discovery endpoints

### Infrastructure

- `internal/infra/orm/` — ent database client
- `internal/infra/cache/` — Redis cache
- `internal/infra/jwt/` — JWT token signing
- `internal/infra/config/` — YAML config loading
- `internal/infra/tracing/` — OpenTelemetry tracing

### Middleware Groups (User API)

| Group | Auth | Purpose |
|---|---|---|
| `publicGroup` | None | Public endpoints (login, register) |
| `authGroup` | JWT | Requires login (profile) |
| `verifiedGroup` | JWT + email verified | Business endpoints |
