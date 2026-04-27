# Card 0644

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/config
Owned Files: internal/config/manager.go, internal/config/config_test.go
Depends On: tasks/cards/done/0643-internal-semver-prerelease-order.md

Goal:
Return errors instead of panicking or silently wrapping when config unmarshal
sets narrow numeric fields.

Scope:
- Check integer, unsigned integer, and float overflow before reflection writes.
- Preserve existing parsing behavior for valid values.
- Add focused tests for int8/uint8/float32 overflow paths.

Non-goals:
- Do not change config source loading.
- Do not change global config helpers.
- Do not change duration parsing behavior.

Files:
- internal/config/manager.go
- internal/config/config_test.go

Tests:
- go test ./internal/config
- go test ./internal/...

Docs Sync:
- None; behavior becomes safer without API changes.

Done Definition:
- Numeric overflow returns an error from `Unmarshal`.
- No reflection panic is possible for configured numeric overflow cases.
- Focused and internal package tests pass.

Outcome:
- Added overflow checks before reflection writes for int, uint, and float fields.
- Added regression coverage for int8, uint8, and float32 overflow.
- Validation: `go test ./internal/config`; `go test ./internal/...`.
