# Card 2040

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: reference/with-websocket
Owned Files:
- reference/with-websocket/main.go
- reference/with-websocket/internal/app/app.go
- reference/with-websocket/internal/config/config.go
- reference/with-websocket/internal/config/config_test.go
- reference/with-websocket/README.md
- reference/with-websocket/ARCHITECTURE.md
- reference/with-websocket/AGENT_TASKS.md
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
- reference/with-websocket/ARCHITECTURE.md
- reference/with-websocket/AGENT_TASKS.md

## Acceptance Tests

- reference/with-websocket/internal/config/config_test.go: TestLoadPrecedenceDefaultsEnvFileEnvironmentFlags
- reference/with-websocket/internal/config/config_test.go: TestLoadIgnoresUnrelatedFlags

## Tests

- Existing package tests.

## Docs Sync

- reference/with-websocket/README.md
- reference/with-websocket/ARCHITECTURE.md
- reference/with-websocket/AGENT_TASKS.md

## Validation

- cd reference/with-websocket && go test -timeout 20s ./...
- cd reference/with-websocket && go vet ./...

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

- Moved signal ownership to `main.run` and changed `App.Start(ctx)` to use caller-owned cancellation.
- Preserved WebSocket shutdown before HTTP shutdown and now surfaces shutdown errors.
- Replaced `.env` `os.Setenv` loading with an overlay-based config loader using `Defaults < .env < process env < flags`; `WS_SECRET` remains validated and is not exposed as a flag.
- Added config precedence, unrelated flag, and missing `WS_SECRET` tests.
- Validation:
  - `cd reference/with-websocket && go test -timeout 20s ./...`
  - `cd reference/with-websocket && go vet ./...`
  - `gofmt -l reference/with-websocket`
