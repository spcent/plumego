# Card 2113: Devserver HTTP Error Contract

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: medium
State: done
Primary Module: cmd/plumego
Owned Files:
- cmd/plumego/internal/devserver/config_edit.go
- cmd/plumego/internal/devserver/dashboard.go
- cmd/plumego/internal/devserver/deps.go
- cmd/plumego/internal/devserver/dashboard_info_test.go
- cmd/plumego/README.md
Depends On: none

Goal:
- Make local devserver HTTP APIs use a consistent structured error contract.
- Replace lowercase/freeform error codes and raw `err.Error()` message concatenation with stable codes and safe messages.

Scope:
- Audit config editing, dashboard API, and dependency-inspection handlers.
- Add small local helpers only if they eliminate duplicated write-error boilerplate.
- Preserve devserver-only behavior and UI endpoints.
- Add focused tests around representative config, dashboard, and dependency error responses.

Non-goals:
- Do not change generated code output, CLI flags, or embedded UI assets.
- Do not move devserver code into stable roots or `x/*`.
- Do not alter proxy, pprof, runner, or alert behavior unless tests require a response assertion update.

Files:
- `cmd/plumego/internal/devserver/config_edit.go`: normalize config endpoint errors.
- `cmd/plumego/internal/devserver/dashboard.go`: normalize dashboard endpoint errors.
- `cmd/plumego/internal/devserver/deps.go`: normalize dependency endpoint errors.
- `cmd/plumego/internal/devserver/dashboard_info_test.go`: add or update HTTP error assertions.
- `cmd/plumego/README.md`: document devserver API error behavior only if user-facing behavior changes.

Tests:
- `go test -race -timeout 60s ./cmd/plumego/...`
- `go test -timeout 20s ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- Required only if public CLI/devserver API behavior or documented local UI behavior changes.

Done Definition:
- Devserver endpoints in scope use uppercase stable error codes.
- Client-facing messages do not include raw filesystem, parser, or toolchain error strings by default.
- Focused tests assert representative error bodies.
- The three listed validation commands pass.

Outcome:
- Added local devserver error codes and a shared `writeDevserverError` helper for config, dashboard, pprof, metrics, deps, build, and API-test failures.
- Replaced raw filesystem, parser, app, pprof, dependency, and toolchain error text in default client-facing responses with safe messages.
- Updated devserver event subscription calls to the current `x/pubsub` context-aware API because `cmd/plumego` validation was otherwise blocked.
- Added focused tests for config edit read failures, dashboard app-not-running responses, and dependency graph failures.
- Documented devserver API error behavior in `cmd/plumego/README.md`.
- Validation note: `cmd/plumego` is an independent Go module, so validation was run from `cmd/plumego` with `GOCACHE=/tmp/plumego-go-cache`.
- Validation passed:
  - `GOCACHE=/tmp/plumego-go-cache go test -race -timeout 60s ./...`
  - `GOCACHE=/tmp/plumego-go-cache go test -timeout 20s ./...`
  - `GOCACHE=/tmp/plumego-go-cache go vet ./...`
