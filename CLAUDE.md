## Project Overview

Go project template — a minimal API project with User model, JWT auth, and admin panel.

## Common Commands

### Build & Run

```bash
go build ./...                # Build
go generate ./...             # Run code generation (ent ORM + autodi wiring)
go run . serve api            # Start User API server
go run . serve admin          # Start Admin API server
go run . eventbus             # Start event bus consumer (Kafka)
go run . worker               # Start background job worker (Asynq/Redis)
go run . scheduler            # Start cron job scheduler
go run . scheduler -l         # List all registered cronjobs
go run . scheduler --once     # Run all jobs once and exit
go run . eventbus topic create # Create Kafka topics
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
- `internal/infra/cache/` — Redis cache + distributed locking
- `internal/infra/jwt/` — JWT token signing
- `internal/infra/config/` — YAML config loading
- `internal/infra/tracing/` — OpenTelemetry tracing
- `internal/infra/bus/` — Event bus (Kafka producer/consumer, DLQ, pattern matching)
- `internal/infra/queue/` — Job queue (Asynq/Redis, retry, tracking)
- `internal/infra/cron/` — Cron scheduler (robfig/cron)

### Event-Driven Architecture

- **Event Bus** (`internal/infra/bus/`): Kafka-based pub/sub with pattern matching, DLQ, and middleware
- **Listeners** (`internal/listener/`): Implement `bus.Listener` interface to react to events
- **Job Queue** (`internal/infra/queue/`): Redis-backed async job processing with retry, tracking, and idempotency
- **Job Handlers** (`internal/job/`): Implement `job.Handler` interface for background tasks
- **Scheduler** (`pkg/schedule/`): Cron-based task scheduling with distributed locking and overlap prevention
- **CronJobs** (`internal/cronjob/`): Implement `schedule.CronJob` interface for periodic tasks

### Middleware Groups (User API)

| Group | Auth | Purpose |
|---|---|---|
| `publicGroup` | None | Public endpoints (login, register) |
| `authGroup` | JWT | Requires login (profile) |
| `verifiedGroup` | JWT + email verified | Business endpoints |
