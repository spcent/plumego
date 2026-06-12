# mini-saas-api

A multi-tenant SaaS API use-case built on Plumego.

Demonstrates stable roots (`core`, `contract`, `middleware`, `security`,
`store`, `health`, `log`, `metrics`) plus beta extensions `x/rest`,
`x/tenant`, and `x/observability` — explicitly wired, no framework magic.

## Run

```bash
cp env.example .env           # edit APP_JWT_SECRET before running in production
go run .                      # listens on :8090
```

Walk the whole API with the bundled examples:

```bash
bash api/curl.sh              # end-to-end walkthrough against a running server
# or open api/examples.http in your editor (REST Client / JetBrains)
# or import api/postman_collection.json into Postman
```

## What it demonstrates

| Concern | Built with |
|---|---|
| App shape, lifecycle, graceful shutdown | `core` (canonical `reference/standard-service` layout) |
| Middleware chain (requestid → headers → CORS → recovery → accesslog → bodylimit → metrics → timeout) | `middleware/*` |
| Signup/login, bcrypt hashing, password strength | `security/password`, `security/authn` |
| Access tokens (HS256 from `APP_JWT_SECRET`), per-route guards | `security/jwt` |
| Single-use refresh tokens, rotation, family revocation on reuse | app `session` domain over `store/kv` |
| Brute-force guard on auth endpoints | `middleware/abuseguard` |
| Tenant resolution, per-tenant rate limit + request quota | `x/tenant` (beta) |
| Simplified RBAC (owner > admin > member), last-owner invariant | app `access` domain, enforced per route |
| Projects CRUD resource controller | `x/rest` (beta) |
| Idempotency-Key replay protection on mutating routes | stable `store/idempotency` contract |
| Tenant-scoped append-only audit log | app `audit` domain |
| Prometheus metrics on `/metrics` (token-guardable) | `x/observability` (beta) |
| Liveness/readiness | `health` |

Persistence is in-memory behind repository interfaces by design (the embedded
`store/kv` file holds JWT keys and refresh tokens). Swapping in a database
means implementing the repository interfaces — no handler changes.

## Routes

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | /healthz | — | Liveness probe |
| GET | /readyz | — | Readiness probe |
| POST | /api/v1/auth/signup | — | Create user + tenant; returns token pair |
| POST | /api/v1/auth/login | — | Email + password → token pair |
| POST | /api/v1/auth/refresh | — | Rotate refresh token |
| GET | /api/v1/me | JWT | Current principal |
| GET | /api/v1/tenant | JWT member | Workspace details |
| PATCH | /api/v1/tenant | JWT admin | Rename workspace |
| GET | /api/v1/tenant/members | JWT member | List members |
| POST | /api/v1/tenant/members | JWT admin | Add member |
| PATCH | /api/v1/tenant/members/:id | JWT admin | Change role |
| DELETE | /api/v1/tenant/members/:id | JWT admin | Remove member |
| GET | /api/v1/tenant/audit?limit=N | JWT admin | Audit log, newest first |
| GET | /api/v1/projects | JWT member | List projects |
| POST | /api/v1/projects | JWT member | Create project |
| GET | /api/v1/projects/:id | JWT member | Get project |
| PUT | /api/v1/projects/:id | JWT member | Update project |
| DELETE | /api/v1/projects/:id | JWT admin | Delete project |
| GET | /metrics | optional token | Prometheus metrics |

All mutating `/api/v1` routes accept an `Idempotency-Key` header: retries with
the same key and payload replay the stored response (look for
`X-Idempotency-Replay: true`); the same key with a different payload is
rejected.

## Configuration

See `env.example` for all `APP_*` variables.
`APP_JWT_SECRET` must be at least 32 characters; the server refuses to start
otherwise. Auth notes:

- Access tokens expire after `APP_JWT_ACCESS_TTL` (default 15 m); refresh via
  `POST /api/v1/auth/refresh`.
- Refresh tokens are single-use. Reusing a consumed token revokes the whole
  session family (token-theft containment) — the client must log in again.
- Role changes take effect when the access token is refreshed or re-issued.

## Module Milestone

`tasks/milestones/active/M-025-mini-saas-api/`
