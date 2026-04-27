# Card 0158

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/config.go`
- `core/app.go`
- `core/options.go`
- `core/options_test.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Normalize router behavior policy into typed `AppConfig` so `core.Option`
  remains a dependency injection surface instead of a mixed config bag.

Problem:
- `core` now documents config-first construction with typed `AppConfig`.
- `WithLogger(...)` fits that contract because it injects a non-config runtime
  dependency.
- `WithMethodNotAllowed(...)` does not: it changes owned router behavior and is
  therefore application configuration, not dependency injection.
- Keeping router policy in `Option` leaves `core.New(cfg, ...)` with two
  different construction models at once.

Scope:
- Move method-not-allowed policy into typed `AppConfig`.
- Remove `WithMethodNotAllowed(...)` from the `core` option surface.
- Keep `Option` for true non-config dependencies such as logger injection.
- Update tests and module docs to reflect the normalized contract.

Non-goals:
- Do not reintroduce raw router replacement or mutation hooks.
- Do not redesign the router package option surface.
- Do not add new config helper functions.

Files:
- `core/config.go`
- `core/app.go`
- `core/options.go`
- `core/options_test.go`
- `docs/modules/core/README.md`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./...`
- `go vet ./...`

Docs Sync:
- Keep `core` docs explicit that behavior policy lives in typed config and
  `Option` is only for non-config dependencies.

Done Definition:
- Router behavior policy is owned by typed config, not by `Option`.
- `core.Option` no longer mixes dependency injection with behavioral config.
- Tests and docs reflect the config-first construction contract.

Outcome:
- Moved owned router method-mismatch policy into typed `AppConfig`.
- Removed `core.WithMethodNotAllowed(...)` and the matching app shadow state.
- Kept `core.Option` focused on non-config dependency injection such as logger
  wiring.
- Updated tests and core module docs to reflect the normalized config-first
  contract.
