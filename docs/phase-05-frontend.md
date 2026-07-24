# Phase 5 — Frontend (Customer Portal & Agent Dashboard)

**Goal:** React SPA with separate customer and agent experiences, wired to the Go API and WebSockets, covering MVP UI subtasks from the product brief.

**Depends on:** Phase 3 (API+WS). Phase 4 endpoints should be integrated when present (canned replies, tags, timeline, search).  
**Exit criteria:** Customer can sign up, create a ticket with attachment, reply, see realtime updates; agent can filter queue, assign, internal note, escalate, see SLA timers and timeline; role-aware nav; optimistic replies with rollback.

---

## Stack

| Lib | Use |
|-----|-----|
| Vite + React + TypeScript | App shell |
| Tailwind CSS + shadcn/ui | UI |
| TanStack Query | Server state (tickets, comments) |
| Zustand | Filters, sidebar, realtime connection flags |
| React Router | Portal vs dashboard routes |

App root: `web/`.

---

## App structure (suggested)

```
web/src/
  app/                 # router, providers
  features/
    auth/
    tickets/
    agents/
    realtime/
  shared/              # api client, ui primitives, hooks
  pages/
    customer/
    agent/
    admin/             # minimal if needed
```

---

## Surfaces

### Customer portal

- Signup / login
- Ticket list (own) + empty/loading/error states
- Create ticket form (title, description, category, priority) + attachment
- Ticket detail: public thread, reply form, SLA/status badges
- Realtime: invalidate/update Query cache on WS events

### Agent dashboard

- Login (agent/admin)
- Queue with filters in **URL query params** (status, priority, assignee, team, tag, q)
- Ticket detail: public + internal notes, assign, status controls, escalate
- SLA countdown in list rows and detail
- Canned replies insert into composer
- Tags + saved filters
- Timeline panel (incident-like)
- Search box hitting FTS endpoint

### Shared

- Role-aware nav and button visibility
- Auth token storage + refresh interceptor
- Validated forms (e.g. zod + react-hook-form)
- Optimistic comment create with rollback on error

---

## API client

- Base URL from `import.meta.env`
- Envelope unwrap: `{ success, data, error }`
- 401 → refresh → retry once → logout

---

## Realtime integration

- Connect to `/api/v1/ws` after auth
- On `ticket.*` / `comment.*`: `queryClient.setQueryData` / `invalidateQueries`
- Respect visibility: customers must not render internal-only payloads (server should already filter)

---

## Implementation checklist

- [ ] Vite app scaffold + Tailwind + shadcn
- [ ] Auth pages + token/refresh flow
- [ ] Customer layout + ticket list/create/detail
- [ ] Agent layout + filtered queue (URL params)
- [ ] Comments with optimistic UI
- [ ] Attachments upload UI
- [ ] SLA timers
- [ ] WS → TanStack Query bridge (Zustand for connection state)
- [ ] Timeline, canned replies, tags, saved filters, search
- [ ] Empty / loading / error states polished
- [ ] Basic Playwright smoke (login + create ticket) — can finalize in phase 6

---

## Docs shipped this phase

- [ ] `web/README.md` (dev server, env vars)
- [ ] Root README: how to run API + web together
- [ ] Screenshot or short GIF optional for polish phase

---

## Design note

Follow existing product brief UX goals (clear portals, SLA visibility). This is an internal tool UI — prefer clarity over marketing-landing aesthetics. Reuse shadcn patterns consistently; do not invent a second design system.
