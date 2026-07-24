# Phase 1 — Auth & RBAC

**Goal:** Customers can register/login; agents and admins can login; JWT access + rotating refresh tokens; role middleware gates routes.

**Depends on:** Phase 0  
**Exit criteria:** Register customer → login → refresh → logout works; agent/admin bootstrap users exist; `RequireRoles` rejects cross-role access on a sample protected route; Bruno auth folder green.

---

## Domain

### Entities

- `User` — `ID`, `Email`, `PasswordHash`, `Role` (`customer`|`agent`|`admin`), `Status`, timestamps
- `RefreshToken` — hashed token, `ExpiresAt`, `RevokedAt`, `ReplacedBy`

### Sentinel errors

Reuse wallet style: `ErrNotFound`, `ErrConflict`, `ErrForbidden`, plus `ErrInvalidCredentials`, `ErrUnauthorized`.

### Repository ports

- `UserRepository` — create, get by ID/email, list agents (for later)
- `RefreshTokenRepository` — create, get by hash, revoke, rotate helpers

---

## Migrations

`000002_auth` (name flexible):

- `users` (unique email, role check constraint)
- `refresh_tokens`
- Indexes on email, refresh token hash

Seed strategy (composition root or migration):

- Admin from config (`admin.email` / `admin.password`)
- Sample agent for local demo

---

## Use cases

`AuthUseCase`:

| Method | Behavior |
|--------|----------|
| `Register` | Role forced to `customer`; Argon2id hash; conflict on email |
| `Login` | Verify hash; issue access JWT + refresh; store refresh hash |
| `Refresh` | Rotate refresh (revoke old, link `replaced_by`); reuse detection fails closed |
| `Logout` | Revoke refresh token |

Password encoding format (match wallet):

```text
argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
```

JWT claims: `user_id`, `email`, `role`, `iss`, `exp`, …

---

## HTTP

| Method | Path | Auth | Notes |
|--------|------|------|-------|
| POST | `/api/v1/auth/register` | Public | Customer only |
| POST | `/api/v1/auth/login` | Public | Any role |
| POST | `/api/v1/auth/refresh` | Public | Body refresh token |
| POST | `/api/v1/auth/logout` | Public | Revoke refresh |
| GET | `/api/v1/me` | Bearer | Current user profile |

Middleware:

- `RequireAuth` — parse Bearer, set context `user_id`, `email`, `role`
- `RequireRoles(...Role)` — 403 if mismatch

DTOs with Gin `binding` tags; map domain errors → 400/401/403/404/409.

---

## sqlc queries (examples)

- `CreateUser`, `GetUserByEmail`, `GetUserByID`
- `CreateRefreshToken`, `GetRefreshTokenByHash`, `RevokeRefreshToken`, `ReplaceRefreshToken`

Adapters map `sql.ErrNoRows` → `entity.ErrNotFound`, unique violations → `ErrConflict`.

---

## Implementation checklist

- [ ] Migration `users` + `refresh_tokens`
- [ ] Domain entities + constructors
- [ ] Repository ports + sqlc queries + postgres adapters
- [ ] Argon2 helpers + JWT package wiring
- [ ] `AuthUseCase`
- [ ] Handlers + DTOs + router groups
- [ ] Auth middleware
- [ ] Bootstrap admin/agent on API startup
- [ ] Unit tests: password hash/verify; optional refresh rotation test
- [ ] Bruno: register, login, refresh, me, logout

---

## Docs shipped this phase

- [ ] Update this checklist as items land
- [ ] Bruno auth requests
- [ ] README: default seeded accounts (dev only)

---

## Out of scope here

- Rate limiting login (phase 4)
- Agent invite emails
- OAuth / SSO
