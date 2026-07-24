# Phase 2 — Tickets, Comments & Assignment

**Goal:** Full ticket lifecycle with RBAC: customers create/track/reply publicly; agents assign, filter, add internal notes, change status via state machine; status history recorded.

**Depends on:** Phase 1  
**Exit criteria:** Customer creates ticket and public reply; agent assigns, adds internal note, transitions status; invalid transitions return 400; customer APIs never leak internal notes; post-commit `EventPublisher` hook exists (even if WS is no-op until phase 3).

---

## Domain

### Entities

- `Team`, `TeamMember`
- `Ticket` — title, description, category, priority, status, `CustomerID`, `AssigneeID?`, `TeamID?`
- `Comment` — `TicketID`, `AuthorID`, `Body`, `Visibility` (`public`|`internal`), `DeletedAt?`
- `TicketStatusHistory` — ticket, actor, from/to status, reason, at

### State machine

Implement `Ticket.TransitionTo(next Status) error` with allowed edges from [implementation-plan.md](./implementation-plan.md). Reject everything else with a domain sentinel (e.g. `ErrInvalidTransition`).

### Repository ports

- `TeamRepository`
- `TicketRepository` — create, get, list (scoped filters), update assignment/status/priority fields
- `CommentRepository` — create, list by ticket (filter visibility), soft-delete
- `StatusHistoryRepository` — append, list by ticket

### Event publisher (stub)

Define:

```go
type EventPublisher interface {
    Publish(ctx context.Context, event DomainEvent) error
}
```

Call **after** `WithTransaction` succeeds (not inside a rolled-back tx). Phase 3 binds this to Redis/WS; phase 2 may use a no-op or in-memory logger publisher.

---

## Migrations

Suggested files:

- `000003_teams` — teams, team_members
- `000004_tickets` — tickets + FKs + indexes (customer_id, assignee_id, status, priority, created_at)
- `000005_comments` — comments + soft-delete
- `000006_status_history` — append-only history

Check constraints for status/priority/category enums.

Seed: one demo team; optionally assign sample agent to team.

---

## Use cases

`TicketUseCase` (split files OK if large):

| Method | Who | Behavior |
|--------|-----|----------|
| `Create` | customer (+ agent/admin acting?) | Set status `open`; history row; event `ticket.created` |
| `List` | customer: own only; agent/admin: filters | Filters: status, priority, assignee, team, category |
| `Get` | owner or agent/admin | 404 if customer crosses tenant |
| `UpdateStatus` | agent/admin (customer limited?) | State machine + history + event |
| `Assign` | agent/admin | Set assignee and/or team; history/audit; event |
| `AddComment` | owner public; agent public or internal | Soft rules; event `comment.created` |
| `ListComments` | strip `internal` for customers | |

Customer status updates: either disallow or allow only limited transitions (prefer **agent-driven** status for MVP clarity; customers reply only).

---

## HTTP

| Method | Path | Roles |
|--------|------|-------|
| POST | `/api/v1/tickets` | customer, agent, admin |
| GET | `/api/v1/tickets` | all (scoped) |
| GET | `/api/v1/tickets/:id` | scoped |
| PATCH | `/api/v1/tickets/:id` | agent, admin (status/fields) |
| POST | `/api/v1/tickets/:id/assign` | agent, admin |
| POST | `/api/v1/tickets/:id/comments` | scoped |
| GET | `/api/v1/tickets/:id/comments` | scoped |
| GET | `/api/v1/teams` | agent, admin |
| POST | `/api/v1/teams` | admin |
| GET | `/api/v1/agents` | agent, admin |

Pagination: limit/offset or cursor; include total when cheap.

---

## Implementation checklist

- [ ] Migrations for teams, tickets, comments, status_history
- [ ] Domain entities + `TransitionTo` + tests
- [ ] sqlc queries + postgres adapters
- [ ] `TicketUseCase` with tx + post-commit publish
- [ ] Handlers/DTOs/router + RBAC
- [ ] Internal comment filtering on customer responses
- [ ] No-op/logging `EventPublisher`
- [ ] Integration or HTTP tests: lifecycle + invalid transition + RBAC
- [ ] Bruno ticket folder

---

## Docs shipped this phase

- [ ] Keep this checklist current
- [ ] Document status transition table in README or here once stable

---

## Out of scope here

- SLA due dates (phase 3)
- Attachments (phase 3)
- WebSocket delivery (phase 3)
- Tags, FTS, canned replies (phase 4)
