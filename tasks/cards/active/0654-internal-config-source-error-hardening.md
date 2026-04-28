# Card 0654

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: internal/config
Owned Files: internal/config/source.go, internal/config/global.go, internal/config/env_test.go
Depends On:

Goal:
Make internal config sources fail explicitly on caller cancellation and environment mutation errors.

Scope:
- Return context cancellation errors from environment and file source load entrypoints before doing work.
- Return `os.Setenv` errors from `LoadEnvFile` / `LoadEnv` instead of silently ignoring them.
- Add focused tests for invalid env keys and canceled source loads.

Non-goals:
- Do not remove the existing panic-compatible constructors.
- Do not change public config accessor behavior.
- Do not add TOML support or alter watch semantics.

Files:
- internal/config/source.go
- internal/config/global.go
- internal/config/env_test.go

Tests:
- go test -timeout 20s ./internal/config
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; this preserves the existing API and only makes existing errors observable.

Done Definition:
- `LoadEnvFile` returns an error when an env assignment cannot be applied.
- `EnvSource.Load` and `FileSource.Load` fail with the supplied context error when called with a canceled context.
- Focused tests and internal package validation pass.

Outcome:
