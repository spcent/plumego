# Card 2043

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/with-gateway
Owned Files:
- reference/with-gateway/main.go
- reference/with-gateway/internal/app/app.go
- reference/with-gateway/README.md
- reference/with-gateway/ARCHITECTURE.md
- reference/with-gateway/AGENT_TASKS.md
Depends On: 2042

## Goal

Align `reference/with-gateway` with the canonical startup lifecycle.

## Scope

Move signal ownership to `main.run`, change `App.Start(ctx)` to accept caller-owned cancellation, and update current local docs that describe the startup shape.

## Non-goals

- Do not change gateway proxy behavior.
- Do not change `x/gateway`.
- Do not change route contracts or upstream configuration semantics.

## Files

- reference/with-gateway/main.go
- reference/with-gateway/internal/app/app.go
- reference/with-gateway/README.md
- reference/with-gateway/ARCHITECTURE.md
- reference/with-gateway/AGENT_TASKS.md

## Acceptance Tests

- Documentation grep has no current `App.Start()` lifecycle examples for `reference/with-gateway`.

## Tests

- Existing package tests.

## Docs Sync

- reference/with-gateway/README.md
- reference/with-gateway/ARCHITECTURE.md
- reference/with-gateway/AGENT_TASKS.md

## Validation

- cd reference/with-gateway && go test -timeout 20s ./...
- cd reference/with-gateway && go vet ./...
- gofmt -l reference/with-gateway

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l reference/with-gateway produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome
