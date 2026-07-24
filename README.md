# Customer Support & SLA Desk

Go backend for customer tickets, agent queues, SLA tracking, and support operations.

## Quick start

1. Copy `.env.example` values as needed and start dependencies: `docker compose up -d`.
2. Install sqlc once: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`.
3. Generate database bindings and run: `make sqlc && make run-api`.
4. API health: `GET http://localhost:8080/health`; API routes use `/api/v1`.

The development API auto-applies migrations. Run the worker separately with
`make run-worker` to mark unpaused, unresolved tickets whose SLA is overdue.

Development bootstrap accounts are `admin@example.com` and `agent@example.com`;
change their passwords and `jwt.secret` before any non-local deployment.

## Core API

- Auth: register, login, refresh, logout, and `/me`.
- Tickets: create, list/search, read, update/assign/escalate, and public or internal comments.
- Internal inbound-email mock: `POST /api/v1/internal/email-to-ticket` with `X-Internal-Token`.

Customer requests are scoped to their tickets, and internal comments are omitted
from customer responses. Ticket status transitions are `open ↔ pending`,
`open|pending → resolved`, and `resolved → open|closed`. Pending tickets pause
their SLA; resuming restores the captured remaining duration.

## Layout

`internal/domain` holds rules, `internal/usecase` orchestrates transactions,
and `internal/adapter/repository/postgres` provides persistence. Migrations are
under `migrations/app`; sqlc query sources are under `db/queries`.
