# Phase 0 — Scaffold

**Goal:** Empty repo becomes a runnable Go module with DDD package layout, Postgres + Redis via Docker, sqlc toolchain, and shared packages copied/adapted from digital-wallet-go patterns.

**Exit criteria:** `docker compose up -d` starts Postgres/Redis; `make sqlc` / `make migrate` / `make run-api` work; `GET /health` returns 200; `docs/architecture.md` matches the tree.

---

## Deliverables

### Repository layout

Create the tree described in [architecture.md](./architecture.md):

- `cmd/api`, `cmd/worker`, `cmd/migrate` (stubs OK; worker can no-op)
- `internal/domain/{entity,repository}` (placeholder package docs)
- `internal/usecase` (empty)
- `internal/adapter/http/v1/{handler,dto,router,middleware}`
- `internal/adapter/repository/postgres` (+ `sqlcgen` output dir)
- `internal/infrastructure/{config,database,redis}`
- `pkg/{jwt,logger,migration,response,uuidv7}`
- `migrations/app/`, `db/queries/`, `config/`, `bruno/`, `web/` (optional empty placeholder)
- `sqlc.yaml`, `Makefile`, `docker-compose.yml`, `.env.example`, root `README.md`

### Module & stack

- Go module: `github.com/XoDeR/customer-support-desk-go`
- Go version aligned with digital-wallet-go (1.24+)
- Dependencies: Gin, pgx/v5, redis client, yaml, jwt, argon2 (may land in phase 1), uuid

### sqlc

`sqlc.yaml` example shape:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/queries"
    schema: "migrations/app"
    gen:
      go:
        package: "sqlcgen"
        out: "internal/adapter/repository/postgres/sqlcgen"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_empty_slices: true
```

Add a minimal health/schema migration (e.g. `000001_init` with `schema_migrations` handled by custom migrator like wallet).

### Database & transactions

- `database.NewPostgresConnection` using pgx stdlib or pgxpool + `database/sql`
- `TransactionManager.WithTransaction` storing tx in context
- `BaseRepository` / helper: resolve `*sqlcgen.Queries` with or without `WithTx`

### HTTP skeleton

- Router mounts `/api/v1` and `/health`
- Middleware: request ID, request logger
- Response helper matching wallet envelope

### Docker

`docker-compose.yml`:

- `postgres:16` with app + test DBs
- `redis:7`
- Volume for local attachment storage (used in phase 3)

### Makefile targets

| Target | Action |
|--------|--------|
| `sqlc` | `sqlc generate` |
| `migrate` | `go run ./cmd/migrate` |
| `run-api` | `go run ./cmd/api` |
| `run-worker` | `go run ./cmd/worker` |
| `test` | `go test ./...` |
| `tidy` | `go mod tidy` |

---

## Implementation checklist

- [ ] `go.mod` / module path
- [ ] Package directories + empty `doc.go` where useful
- [ ] Config load (YAML + env) — `config/app.dev.yaml`
- [ ] Logger init
- [ ] Postgres connection + migration manager
- [ ] sqlc config + first empty/health query if needed
- [ ] Gin engine + health handler
- [ ] docker-compose + `.env.example`
- [ ] Makefile
- [ ] Root README (clone, compose, make targets, layout overview)
- [ ] Confirm [architecture.md](./architecture.md) paths match disk

---

## Docs shipped this phase

- [x] `docs/architecture.md`
- [x] `docs/implementation-plan.md`
- [x] `docs/phase-00-scaffold.md` (this file)
- [ ] Root `README.md` setup section

---

## Notes / pitfalls

- Keep sqlc **out** of `internal/domain` and `internal/usecase`.
- Prefer copying wallet’s migration manager and response/JWT packages and adapting imports rather than inventing a new style.
- Do not block phase 0 on frontend tooling; `web/` can wait until phase 5.
