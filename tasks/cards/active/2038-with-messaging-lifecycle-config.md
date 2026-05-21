# Card 2038

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/with-messaging
Owned Files:
- reference/with-messaging/main.go
- reference/with-messaging/internal/app/app.go
- reference/with-messaging/internal/config/config.go
- reference/with-messaging/internal/config/config_test.go
- reference/with-messaging/README.md
Depends On: 2037

## Goal

Align `reference/with-messaging` with the canonical startup lifecycle and config precedence.

## Scope

Move signal ownership to `main.run`, change `App.Start(ctx)` to accept caller-owned cancellation, and make config precedence `Defaults < .env < process env < flags` without mutating process env from `.env`.

## Non-goals

- Do not change messaging broker behavior.
- Do not add production hardening beyond the existing demo surface.
- Do not change route contracts.

## Files

- reference/with-messaging/main.go
- reference/with-messaging/internal/app/app.go
- reference/with-messaging/internal/config/config.go
- reference/with-messaging/internal/config/config_test.go
- reference/with-messaging/README.md

## Acceptance Tests

- reference/with-messaging/internal/config/config_test.go: TestLoadPrecedenceDefaultsEnvFileEnvironmentFlags
- reference/with-messaging/internal/config/config_test.go: TestLoadIgnoresUnrelatedFlags

## Tests

- Existing route and handler tests.

## Docs Sync

- reference/with-messaging/README.md

## Validation

- cd reference/with-messaging && go test -timeout 20s ./...
- cd reference/with-messaging && go vet ./...

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

