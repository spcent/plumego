# Card 2040

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/with-websocket
Owned Files:
- reference/with-websocket/main.go
- reference/with-websocket/internal/app/app.go
- reference/with-websocket/internal/config/config.go
- reference/with-websocket/internal/config/config_test.go
- reference/with-websocket/README.md
Depends On: 2039

## Goal

Align `reference/with-websocket` with the canonical startup lifecycle and config precedence while preserving WebSocket shutdown order.

## Scope

Move signal ownership to `main.run`, change `App.Start(ctx)` to accept caller-owned cancellation, preserve WebSocket shutdown before HTTP shutdown, and make config precedence `Defaults < .env < process env < flags`.

## Non-goals

- Do not change WebSocket auth, route registration, or message behavior.
- Do not change `x/websocket`.
- Do not weaken `WS_SECRET` validation.

## Files

- reference/with-websocket/main.go
- reference/with-websocket/internal/app/app.go
- reference/with-websocket/internal/config/config.go
- reference/with-websocket/internal/config/config_test.go
- reference/with-websocket/README.md

## Acceptance Tests

- reference/with-websocket/internal/config/config_test.go: TestLoadPrecedenceDefaultsEnvFileEnvironmentFlags
- reference/with-websocket/internal/config/config_test.go: TestLoadIgnoresUnrelatedFlags

## Tests

- Existing package tests.

## Docs Sync

- reference/with-websocket/README.md

## Validation

- cd reference/with-websocket && go test -timeout 20s ./...
- cd reference/with-websocket && go vet ./...

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

