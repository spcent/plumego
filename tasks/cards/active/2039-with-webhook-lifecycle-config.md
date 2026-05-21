# Card 2039

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/with-webhook
Owned Files:
- reference/with-webhook/main.go
- reference/with-webhook/internal/app/app.go
- reference/with-webhook/internal/config/config.go
- reference/with-webhook/internal/config/config_test.go
- reference/with-webhook/README.md
Depends On: 2038

## Goal

Align `reference/with-webhook` with the canonical startup lifecycle and config precedence.

## Scope

Move signal ownership to `main.run`, change `App.Start(ctx)` to accept caller-owned cancellation, and make config precedence `Defaults < .env < process env < flags` without mutating process env from `.env`.

## Non-goals

- Do not change webhook signature, pubsub, or inbound delivery behavior.
- Do not add authentication policy beyond the existing demo behavior.
- Do not change route contracts.

## Files

- reference/with-webhook/main.go
- reference/with-webhook/internal/app/app.go
- reference/with-webhook/internal/config/config.go
- reference/with-webhook/internal/config/config_test.go
- reference/with-webhook/README.md

## Acceptance Tests

- reference/with-webhook/internal/config/config_test.go: TestLoadPrecedenceDefaultsEnvFileEnvironmentFlags
- reference/with-webhook/internal/config/config_test.go: TestLoadIgnoresUnrelatedFlags

## Tests

- Existing package tests.

## Docs Sync

- reference/with-webhook/README.md

## Validation

- cd reference/with-webhook && go test -timeout 20s ./...
- cd reference/with-webhook && go vet ./...

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

