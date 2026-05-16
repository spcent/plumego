# Card 1426

Milestone: v1-breaking-normalization
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: log
Owned Files:
- log/*
Depends On:
- 1421

Goal:
- Normalize logger configuration and output semantics before v1 freeze.

Scope:
- Enumerate caller wrapper, level, writer, formatter, and default backend
  compatibility paths before editing.
- Remove redundant caller wrappers or formatter aliases.
- Keep one canonical option shape for level, writer, fields, caller, and output.
- Preserve or intentionally update tests that define v1 output semantics.
- Document any breaking output or caller-depth changes.

Non-goals:
- Do not introduce global mutable registries.
- Do not add non-stdlib logging dependencies.
- Do not change application wiring outside required caller migration.

Files:
- log/logger.go
- log/glog.go
- log/json.go
- log/fields.go
- log/*_test.go

Tests:
- go test -timeout 20s ./log
- go vet ./log
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update logging docs if output fields, caller behavior, or construction changes.

Done Definition:
- Redundant compatibility helpers are removed or consolidated.
- Log tests define the final v1 output contract.
- Dependency checks pass.

Outcome:
- Removed the redundant `StructuredLogger.With(key, value)` shortcut and kept
  `WithFields` as the single derived-logger field attachment API.
- Made unknown `LoggerConfig.Format` values deterministic and non-emitting by
  selecting the discard logger instead of silently falling back to text output.
- Updated log docs and stable API snapshot to reflect the v1 logging surface.
- Validation:
  - `go test -timeout 20s ./log`
  - `go vet ./log`
  - `go run ./internal/checks/dependency-rules`
  - `go build ./...`
