# mini-saas-api

A multi-tenant SaaS API use-case built on Plumego.

Demonstrates stable roots (`core`, `contract`, `middleware`, `security`,
`store`, `health`, `log`, `metrics`) plus beta extensions `x/rest`,
`x/tenant`, and `x/observability` — explicitly wired, no framework magic.

> **Status:** Scaffold in place (card 1525 complete). Auth, tenancy, projects,
> and observability are implemented by cards 1526–1532.

## Run

```bash
cp env.example .env           # edit APP_JWT_SECRET before running in production
go run .
```

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
| GET | /api/v1/tenant/audit | JWT admin | Audit log |
| GET | /api/v1/projects | JWT member | List projects |
| POST | /api/v1/projects | JWT member | Create project |
| GET | /api/v1/projects/:id | JWT member | Get project |
| PUT | /api/v1/projects/:id | JWT member | Update project |
| DELETE | /api/v1/projects/:id | JWT admin | Delete project |
| GET | /metrics | optional token | Prometheus metrics |

## Configuration

See `env.example` for all `APP_*` variables.
`APP_JWT_SECRET` must be at least 32 characters; the server refuses to start otherwise.

## Module Milestone

`tasks/milestones/active/M-025-mini-saas-api/`
