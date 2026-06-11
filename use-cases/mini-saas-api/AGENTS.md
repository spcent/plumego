# AGENTS.md - use-cases/mini-saas-api

Operational guide for agents working in `use-cases/mini-saas-api`.

## 1. Purpose

`mini-saas-api` is a production-scale SaaS API use-case that demonstrates
Plumego's stable roots plus the beta extensions `x/rest`, `x/tenant`, and
`x/observability` wired together explicitly without framework magic.

See `tasks/milestones/active/M-025-mini-saas-api/` for the full milestone spec
and plan.

## 2. Minimal Context (per card)

Read in order:

1. Root `AGENTS.md`
2. This file
3. `ARCHITECTURE.md` for ownership and dependency direction questions
4. Touched Go files and their tests
5. Relevant card in `tasks/cards/active/`

## 3. Boundaries

- Layout copies `reference/standard-service` exactly:
  `main.go` → `internal/config` → `internal/app` → `internal/handler` → `internal/domain`
- Handlers must not import `internal/app` or `internal/config`.
- Domain packages must not import handler packages.
- Beta extensions (`x/rest`, `x/tenant`, `x/observability`) are imported only in
  `internal/app/app.go` and `internal/app/routes.go` — never in handler or domain code.
- No new external dependencies in `go.mod` without milestone approval.
- No `init()` registration, no reflection routing, no hidden globals.
- `contract.WriteResponse` / `contract.WriteError` are the only response paths.
- Secrets (JWT secret, password hashes, refresh tokens) must never appear in logs.

## 4. Validation (per card)

```bash
cd use-cases/mini-saas-api && go vet ./... && go test -race -timeout 60s ./...
gofmt -l use-cases/mini-saas-api
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
```
