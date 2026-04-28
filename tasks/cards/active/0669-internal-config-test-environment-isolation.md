# Card 0669

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: internal/config
Owned Files: internal/config/env_test.go, internal/config/global_test.go, internal/config/config_test.go
Depends On:

Goal:
Stop internal config tests from mutating process-wide environment and global config state without automatic restoration.

Scope:
- Replace broad `os.Clearenv` usage with targeted test environment setup.
- Use `t.Setenv` or cleanup helpers for test-owned environment variables.
- Ensure tests that initialize or replace the global config reset it during cleanup.

Non-goals:
- Do not change runtime config behavior.
- Do not rewrite unrelated test assertions.

Files:
- internal/config/env_test.go
- internal/config/global_test.go
- internal/config/config_test.go

Tests:
- go test -timeout 20s ./internal/config
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; test isolation only.

Done Definition:
- Config tests no longer call `os.Clearenv`.
- Test-owned environment variables and global config state are restored by cleanup paths.

Outcome:

