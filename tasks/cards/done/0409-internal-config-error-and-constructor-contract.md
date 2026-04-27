# Card 0409: Internal Config Error And Constructor Contract

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: high
State: done
Primary Module: internal/config
Owned Files:
- internal/config/validator.go
- internal/config/manager.go
- internal/config/source.go
- internal/config/config_test.go
Depends On: none

Goal:
- Make internal config validation errors safe for HTTP serialization.
- Add explicit error-returning paths for config constructors/options that currently panic on invalid inputs.

Scope:
- Audit `ValidateAll` error construction for raw validation text and raw config value leakage.
- Add or use safe, stable messages for required and validation-failed config errors.
- Add error-returning variants for nil logger manager construction and invalid file watch interval while preserving panic compatibility for existing callers.
- Add focused tests for safe validation details and new explicit error paths.

Non-goals:
- Do not change config source precedence, watching semantics, environment parsing, or global config behavior.
- Do not introduce external dependencies.
- Do not expose `internal/config` as a stable public package.

Files:
- `internal/config/validator.go`: remove raw `err.Error()` and raw config values from public error messages/details.
- `internal/config/manager.go`: add explicit error-returning construction for nil logger input while preserving existing panic wrapper.
- `internal/config/source.go`: add explicit error-returning watch interval option while preserving existing panic wrapper.
- `internal/config/config_test.go`: cover safe validation errors and constructor/option compatibility.

Tests:
- `go test -race -timeout 60s ./internal/config`
- `go test -timeout 20s ./internal/config`
- `go vet ./internal/config`

Docs Sync:
- Not required unless repository docs currently describe exact internal config error payloads or constructor panic behavior.

Done Definition:
- Config validation errors no longer expose raw validator messages or raw config values through `contract.APIError`.
- Nil logger and invalid watch interval have explicit error-returning paths.
- Existing panic-based APIs keep compatibility by delegating to the explicit variants.
- The three listed validation commands pass.

Outcome:
- Added `NewManagerE` and `ErrLoggerRequired` for explicit nil-logger construction failures while preserving `NewManager` panic compatibility.
- Added `WithWatchIntervalE` and `ErrInvalidWatchInterval` for explicit invalid watch interval handling while preserving `WithWatchInterval` panic compatibility.
- Changed config schema validation errors to use safe public messages and removed raw config values from `APIError.Details`.
- Added tests for safe schema validation errors, explicit constructor/option errors, and compatibility panics.
- Validation passed:
  - `go test -race -timeout 60s ./internal/config`
  - `go test -timeout 20s ./internal/config`
  - `go vet ./internal/config`
