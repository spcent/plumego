# Card 2037

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Context Package: implementation
Priority: P0
State: done
Primary Module: reference/production-service
Owned Files:
- reference/production-service/main.go
- reference/production-service/internal/app/app.go
- reference/production-service/internal/app/app_test.go
- reference/production-service/README.md
Depends On:

## Goal

Make `reference/production-service` pass its smoke test without weakening `/ops/metrics` authentication.

## Scope

Fix the production reference's ops-token test/config mismatch and align its startup lifecycle with `reference/standard-service`'s `main.run -> app.Start(ctx)` shape.

## Non-goals

- Do not make `/ops/metrics` public.
- Do not change stable root packages or `x/observability`.
- Do not introduce fallback secrets or hard-coded production tokens.

## Files

- reference/production-service/main.go
- reference/production-service/internal/app/app.go
- reference/production-service/internal/app/app_test.go
- reference/production-service/README.md

## Acceptance Tests

- reference/production-service/internal/app/app_test.go: TestProductionServiceSmoke

## Tests

- Confirm `/ops/metrics` still returns `401` without a bearer token.
- Confirm `/ops/metrics` returns `200` with the configured ops token.

## Docs Sync

- reference/production-service/README.md

## Validation

- cd reference/production-service && go test -timeout 20s ./...
- make reference-test

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Completed production-service ops metrics smoke repair while preserving fail-closed
auth. The test now configures the app-local ops token when bypassing
`config.Load`, and the service lifecycle now matches the canonical
`main.run -> app.Start(ctx)` shape.
