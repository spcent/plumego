# AGENTS.md - use-cases/

`use-cases/` contains production-scale application repositories built on Plumego.
They differ from `reference/` in scope and complexity:

| Directory | Purpose |
|---|---|
| `reference/` | Teaching apps — minimal, focused, each demonstrating one capability |
| `use-cases/` | Production-scale apps — full domain logic, external deps, own deployment |

Use-case apps have their own `go.mod` with a `replace` directive pointing to `../..`.
They are NOT canonical application templates. Each has its own `AGENTS.md` with
local operating rules.

## Apps

| App | What it is |
|---|---|
| `dbadmin/` | Local-first multi-database admin workbench — MySQL, SQLite, Redis, MongoDB, and Elasticsearch inspection with safety guardrails |
| `mini-saas-api/` | Multi-tenant SaaS API showcase — JWT auth with rotated refresh tokens, RBAC, x/tenant rate limit + quota, x/rest projects CRUD, idempotent writes, audit log, Prometheus metrics; zero external deps |
| `workerfleet/` | Worker fleet monitoring service — HTTP ingestion, domain events, MongoDB storage, Kubernetes reconciliation, Prometheus metrics, Feishu/webhook alerting |

For canonical application layout, see `reference/standard-service`.
For repository-wide rules, see the root `AGENTS.md`.
