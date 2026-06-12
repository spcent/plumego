# Card 1525

Milestone: M-025
Recipe: (none — new use-case scaffold)
Context Package: use-cases/mini-saas-api
Priority: P1
State: done
Primary Module: use-cases/mini-saas-api
Owned Files: go.mod, main.go, internal/config/config.go, internal/app/app.go, internal/app/routes.go, internal/handler/health.go, env.example, AGENTS.md, ARCHITECTURE.md, README.md (new, all files)
Depends On: (none)

## Goal

`use-cases/mini-saas-api` compiles, passes `go vet`, and its lifecycle tests
pass: `/healthz` returns 200, `/readyz` returns 200, graceful shutdown returns
nil — establishing the canonical app skeleton that subsequent cards extend.

## Scope

New use-case module only; zero changes to `reference/` or stable roots.

Key files in scope:
- use-cases/mini-saas-api/go.mod (new)
- use-cases/mini-saas-api/main.go (new)
- use-cases/mini-saas-api/internal/config/config.go (new)
- use-cases/mini-saas-api/internal/app/app.go (new)
- use-cases/mini-saas-api/internal/app/routes.go (new)
- use-cases/mini-saas-api/internal/handler/health.go (new)
- use-cases/mini-saas-api/internal/handler/write.go (new)
- use-cases/mini-saas-api/env.example (new)
- use-cases/mini-saas-api/AGENTS.md (stub, new)
- use-cases/mini-saas-api/ARCHITECTURE.md (stub, new)
- use-cases/mini-saas-api/README.md (stub, new)
- use-cases/mini-saas-api/internal/config/config_test.go (new)
- use-cases/mini-saas-api/internal/app/app_test.go (new)
- use-cases/mini-saas-api/internal/handler/health_test.go (new)

## Non-goals

- No domain models, auth, tenant, REST CRUD, or observability wiring (those are cards 1526–1532).
- No changes to reference/* or any stable root or x/* package.
- No external dependencies in go.mod.
- No JWT manager construction or store/kv initialization (those come with auth card 1527).

## Files

use-cases/mini-saas-api/go.mod
use-cases/mini-saas-api/main.go
use-cases/mini-saas-api/internal/config/config.go
use-cases/mini-saas-api/internal/app/app.go
use-cases/mini-saas-api/internal/app/routes.go

## Acceptance Tests

internal/app/app_test.go: TestAcceptanceHealthLiveness
internal/app/app_test.go: TestAcceptanceHealthReadiness
internal/app/app_test.go: TestAcceptanceGracefulShutdown

## Tests

internal/config/config_test.go: TestValidateRejectsEmptyAddr
internal/config/config_test.go: TestValidateRejectsShortJWTSecret
internal/config/config_test.go: TestDefaultsPassValidation
internal/app/app_test.go: TestRegisterRoutesShape
internal/app/app_test.go: TestMiddlewareSecurityHeaders
internal/app/app_test.go: TestMiddlewarePanicRecovery
internal/app/app_test.go: TestMiddlewareBodyLimit
internal/handler/health_test.go: TestHealthLiveAlwaysOK
internal/handler/health_test.go: TestHealthReadyNoCheckers
internal/handler/health_test.go: TestHealthReadyCheckerFails

## Docs Sync

use-cases/AGENTS.md — register mini-saas-api in the app table (card 1533).

## Validation

cd use-cases/mini-saas-api && go vet ./... && go test -race -timeout 60s ./...
gofmt -l use-cases/mini-saas-api

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

`use-cases/mini-saas-api` scaffolded with go.mod + replace directive, main.go, config
(JWT secret validated ≥32 chars, access/refresh TTLs, metrics token), full middleware
chain (requestid → securityheaders → cors → recovery → accesslog → bodylimit →
httpmetrics/noop → timeout), /healthz and /readyz routes, health handler, and doc
stubs (AGENTS.md, ARCHITECTURE.md, README.md, env.example). All acceptance tests
pass with -race. Boundary checks (dependency-rules, agent-workflow) pass. gofmt clean.
Card provisional IDs 2070–2077 in plan-M-025.md corrected to 1525–1532.
