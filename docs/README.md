# Customer Support & SLA Desk — Docs

Source product brief: `customer-support-desk-go-docs/customer-support-sla-desk.md`  
Reference architecture: `digital-wallet-go` (DDD modular monolith; this project uses **sqlc** instead of sqlx).

## Index

| Doc | Purpose |
|-----|---------|
| [implementation-plan.md](./implementation-plan.md) | Locked decisions, domain model, API sketch, phase index |
| [architecture.md](./architecture.md) | DDD layers, sqlc wiring, processes, API envelope |
| [phase-00-scaffold.md](./phase-00-scaffold.md) | Toolchain, layout, docker-compose, sqlc |
| [phase-01-auth.md](./phase-01-auth.md) | Signup/login, JWT, RBAC, seeds |
| [phase-02-tickets.md](./phase-02-tickets.md) | Tickets, comments, assignment, status machine |
| [phase-03-sla-attachments-realtime.md](./phase-03-sla-attachments-realtime.md) | SLA timers, files, WebSockets |
| [phase-04-optional-ops.md](./phase-04-optional-ops.md) | Worker, queues, FTS, audit, tags, email mock |
| [phase-05-frontend.md](./phase-05-frontend.md) | React customer portal + agent dashboard |
| [phase-06-polish.md](./phase-06-polish.md) | Tests, README diagrams, demo seed |

## Documentation cadence

Write or update docs **in the same phase as the feature**. Phase files are the implementation checklists; keep them accurate as code lands. Feature deep-dives (auth flows, SLA rules, realtime) live inside the phase doc until split out if a doc grows too large.

## Status legend (in phase checklists)

- `[ ]` not started
- `[~]` in progress
- `[x]` done
