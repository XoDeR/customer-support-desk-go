# Phase 4 — Optional Ops Features

**Goal:** Ship every Optional checklist item from the product brief: SLA breach worker, escalation + notification queue, rate limits, full-text search, audit timeline API, canned replies, tags/saved filters, email-to-ticket mock.

**Depends on:** Phase 3  
**Exit criteria:** Worker marks breached tickets and publishes `sla.breached`; escalate enqueues and worker notifies; rate limits return 429; FTS query returns ranked tickets; timeline endpoint complete; agent canned/tags/filters work; mock email endpoint creates a ticket.

---

## Feature breakdown

### 1. SLA breach worker

- `cmd/worker` cron (e.g. every 1m): select tickets where `sla_due_at < now`, `breached_at IS NULL`, status not `resolved`/`closed`, not paused
- Set `breached_at`, append audit, publish Redis + WS event
- Idempotent: skip if already breached

### 2. Escalation + notification queue

**Escalate API** `POST /api/v1/tickets/:id/escalate`:

- Raise priority one notch (cap `urgent`)
- Recalculate SLA (phase 3 rules)
- Audit event `escalated`
- `LPUSH` / Redis stream job `{ticket_id, actor_id, new_priority}`

**Worker consumer:**

- Pop jobs; log notification; publish WS notify to assignee/team channel
- Dead-letter or retry lightly on failure (keep simple: log + requeue N times)

### 3. Rate limits (Redis sliding window)

| Action | Suggested limit |
|--------|-----------------|
| Ticket create per user | 10 / 10m |
| Reply per user | 60 / 10m |
| Login per IP/email | 10 / 15m |

Middleware or use-case check → `429` + `ErrRateLimited`.

### 4. Full-text search

- `tsvector` on ticket title/description (+ comment bodies via trigger or materialized column)
- GIN index
- sqlc query `SearchTickets(query, filters…)`
- Agent/admin: `GET /api/v1/tickets/search?q=`
- Customers: search restricted to own tickets

### 5. Audit log & timeline

- Unify status/assign/priority/escalate/breach into `audit_events` or extend `ticket_status_history` into a richer `ticket_events` table
- `GET /api/v1/tickets/:id/timeline` returns ordered incident-like entries for UI

### 6. Canned replies

- Table `canned_replies` (title, body, team_id nullable, created_by)
- CRUD for agent/admin
- FE uses them in composer (phase 5)

### 7. Tags & saved filters

- `tags` + `ticket_tags` (or text[] on tickets — prefer normalized)
- Filter tickets by tag
- `saved_filters` — user_id, name, query JSON (status/priority/assignee/tag/…)
- CRUD for agents

### 8. Email-to-ticket mock

`POST /api/v1/internal/email-to-ticket`

Headers: shared secret (`X-Internal-Token`)  
Body example:

```json
{
  "from_email": "user@example.com",
  "subject": "Cannot login",
  "body": "...",
  "category": "account",
  "priority": "medium"
}
```

Behavior: find-or-create customer by email (random temp password unused / must reset later — or require existing user only; **prefer find-or-create customer with unusable password flag** documented in README); create ticket; return id.

---

## Migrations

- `000009_audit_or_events` (if not covered)
- `000010_fts`
- `000011_canned_replies`
- `000012_tags_saved_filters`
- Worker needs no new DB beyond breach column (may already exist)

---

## HTTP additions

| Method | Path | Roles |
|--------|------|-------|
| POST | `/tickets/:id/escalate` | agent, admin |
| GET | `/tickets/search` | scoped |
| GET | `/tickets/:id/timeline` | scoped |
| CRUD | `/canned-replies` | agent, admin |
| CRUD | `/tags`, ticket tag attach | agent, admin |
| CRUD | `/saved-filters` | agent, admin |
| POST | `/internal/email-to-ticket` | secret |

---

## Implementation checklist

- [ ] SLA breach job in `cmd/worker`
- [ ] Escalate use case + Redis queue + consumer
- [ ] Redis rate limiter + wire to create/reply/login
- [ ] FTS migration + search endpoint
- [ ] Timeline/audit API complete
- [ ] Canned replies CRUD
- [ ] Tags + saved filters
- [ ] Email-to-ticket mock
- [ ] Bruno coverage for each feature
- [ ] Worker README section (how to run both binaries)

---

## Docs shipped this phase

- [ ] This checklist updated
- [ ] Rate-limit table finalized vs code
- [ ] Internal email mock contract documented in README

---

## Notes

Optional features are **in scope** for this project (not deferred). Prefer shipping thin vertical slices (API + migration + Bruno) per feature rather than one giant PR if splitting work.
