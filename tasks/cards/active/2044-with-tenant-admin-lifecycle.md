# Card 2044

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/with-tenant-admin
Owned Files:
- reference/with-tenant-admin/main.go
- reference/with-tenant-admin/internal/app/app.go
- reference/with-tenant-admin/README.md
Depends On: 2043

## Goal

Align `reference/with-tenant-admin` with the canonical startup lifecycle.

## Scope

Move signal ownership to `main.run`, change `App.Start(ctx)` to accept caller-owned cancellation, and update current local docs that describe the startup shape.

## Non-goals

- Do not change tenant admin route behavior.
- Do not change `x/tenant` packages.
- Do not change auth, policy, or handler contracts.

## Files

- reference/with-tenant-admin/main.go
- reference/with-tenant-admin/internal/app/app.go
- reference/with-tenant-admin/README.md

## Acceptance Tests

- Documentation/code grep has no `App.Start()` self-signal lifecycle for `reference/with-tenant-admin`.

## Tests

- Existing package tests.

## Docs Sync

- reference/with-tenant-admin/README.md

## Validation

- cd reference/with-tenant-admin && go test -timeout 20s ./...
- cd reference/with-tenant-admin && go vet ./...
- gofmt -l reference/with-tenant-admin

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l reference/with-tenant-admin produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome
