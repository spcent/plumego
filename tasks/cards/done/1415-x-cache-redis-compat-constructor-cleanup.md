# Card 1415

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/cache/redis
Owned Files:
- x/cache/redis/redis.go
- x/cache/redis/redis_test.go
- docs/modules/x-cache/README.md
Depends On:
- 1414

Goal:
- Clarify and constrain `x/cache/redis` compatibility constructors and mutable adapter options.

Scope:
- Inventory compatibility constructor paths and comments.
- Prefer validated construction paths for new code while preserving compatibility constructors.
- Add or tighten tests around mutable compatibility state and unsupported clear behavior.
- Keep Redis client abstraction dependency-free.

Non-goals:
- Do not add a Redis client dependency.
- Do not remove exported compatibility fields or constructors.
- Do not change cache interface contracts.

Files:
- x/cache/redis/redis.go
- x/cache/redis/redis_test.go
- docs/modules/x-cache/README.md

Tests:
- go test -timeout 20s ./x/cache/redis
- go vet ./x/cache/redis
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update `docs/modules/x-cache/README.md` with constructor guidance if behavior text changes.

Done Definition:
- Redis adapter compatibility paths are explicit and tested.
- New-code constructor guidance is clear.
- Redis adapter tests and dependency checks pass.

Outcome:
- Completed on 2026-05-15.
- Added validated capability constructors for counter, appender, and atomic Redis adapters while preserving existing compatibility constructors.
- Centralized compatibility adapter construction for existing optional-capability constructors.
- Added tests for validated capability constructors, invalid config rejection, unsupported capability errors, and frozen option behavior.
- Updated `docs/modules/x-cache/README.md` with constructor guidance.
- Validation:
  - `go test -timeout 20s ./x/cache/redis`
  - `go vet ./x/cache/redis`
  - `go run ./internal/checks/dependency-rules`
