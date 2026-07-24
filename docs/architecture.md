# Architecture

This service is a **DDD / Clean Architecture** modular monolith: Go API + worker binaries, one PostgreSQL database, Redis for fanout / queues / rate limits, and a React SPA under `web/`.

## Layers

| Layer | Location | Responsibility |
|-------|----------|----------------|
| Domain | `internal/domain` | Entities, invariants, repository interfaces, sentinel errors |
| Application | `internal/usecase` | Use cases; orchestrate transactions and domain rules |
| Adapters | `internal/adapter` | HTTP/WS, Postgres (sqlc wrappers), storage, Redis publishers |
| Infrastructure | `internal/infrastructure` | Config, DB, transaction manager, Redis client |
| Shared | `pkg/` | Logger, JWT, UUID v7, migrations, response envelope |
| Frontend | `web/` | Customer portal + agent dashboard |

**Dependency rule:** adapters and infrastructure depend inward on domain/use cases. Domain never imports adapters, infrastructure, or sqlc-generated code.

## Package layout

```
cmd/
  api/main.go              # HTTP + WS composition root
  worker/main.go           # SLA breach + escalation consumer
  migrate/main.go          # migrations only
internal/
  domain/
    entity/
    repository/            # ports only
  usecase/
  adapter/
    http/v1/{handler,dto,router,middleware}
    ws/
    repository/postgres/   # wraps sqlcgen; implements ports
    storage/               # local (+ R2 stub)
  infrastructure/
    config/
    database/              # pgx pool / database/sql + TransactionManager
    redis/
pkg/
  jwt, logger, migration, response, uuidv7
db/
  queries/                 # sqlc .sql sources
migrations/app/            # numbered .up.sql / .down.sql
web/                       # Vite React TypeScript
bruno/                     # API collection (grown per phase)
docs/
sqlc.yaml
```

## Composition root

`cmd/api/main.go` and `cmd/worker/main.go` wire repositories → use cases → handlers/jobs. There is no DI container (same as digital-wallet-go).

## Persistence (sqlc)

- Schema changes live in `migrations/app/`.
- Queries live in `db/queries/*.sql`; `sqlc generate` writes to `internal/adapter/repository/postgres/sqlcgen`.
- Postgres adapters implement `internal/domain/repository` interfaces and call sqlc `Queries`.
- `TransactionManager.WithTransaction` stores `*sql.Tx` (or pgx tx via stdlib) in `context`; repositories use `Queries.WithTx(tx)` when a tx is present.

```text
usecase → domain/repository.TicketRepository
              ↑
adapter/repository/postgres.TicketRepository
              → sqlcgen.Queries (WithTx from context)
```

Use cases **must not** import `sqlcgen`.

## Identifiers

Primary keys are **UUID v7** (time-ordered) for better B-tree locality.

## API shape

- Base path: `/api/v1`
- JSON envelope: `{ success, message, data, error, timestamp }`
- Auth: `Authorization: Bearer <access_token>`
- Realtime: `GET /ws` with the same Bearer token

## Processes

| Binary | Role |
|--------|------|
| `cmd/api` | HTTP API, WebSocket hub, auto-migrate on startup (dev) |
| `cmd/migrate` | Apply migrations only |
| `cmd/worker` | SLA breach cron, escalation queue consumer |

## Realtime & workers

- After a successful DB commit, use cases publish domain events to an `EventPublisher`.
- API process: Redis pub/sub → local WS hub → connected clients.
- Worker process: cron for SLA breaches; Redis list/stream consumer for escalation notifications.

## Non-goals

- Microservices / DB-per-service
- GraphQL
- Live email provider (mock endpoint only)
- Business-hours SLA calendars (wall clock only for v1)
