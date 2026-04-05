# Card 0775

Priority: P1
State: active
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
