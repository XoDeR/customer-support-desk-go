# Web — Customer Support & SLA Desk

Vite + React + TypeScript + Tailwind + TanStack Query + Zustand.

## Run

```bash
npm install
npm run dev
```

Optional env (`.env` in `web/`):

```
VITE_API_URL=http://localhost:8080/api/v1
```

Open http://localhost:5173

### Accounts

- Customer: register at `/register`
- Agent: `agent@example.com` / `agent-password-change-me`
- Admin: `admin@example.com` / `admin-password-change-me`
- Seeded customer (after `make seed`): `customer@example.com` / `customer-password-change-me`

### E2E smoke

With API (`make run-api`) and `npm run dev` running:

```bash
npx playwright install chromium
npm run test:e2e
```

## Features

- Customer portal: tickets, create, public replies, attachments, realtime cache invalidation
- Agent dashboard: URL filters, saved filters, assign/status/escalate, internal notes, canned replies, tags, timeline
- Agent tools page: manage canned replies and tags
