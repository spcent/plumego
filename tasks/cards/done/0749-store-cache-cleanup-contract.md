# Card 0749

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md

Goal:
Make expired cache cleanup behavior deterministic for stable in-process cache use.

Scope:
- Remove the hard-coded 1000-item cleanup cap or make it configurable with a documented default.
- Add tests proving cleanup can remove more than 1000 expired items.
- Sync store docs with cleanup behavior.

Non-goals:
- Do not add a distributed cache cleaner.
- Do not add heap/index data structures.
- Do not change TTL semantics.

Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/cache
- go vet ./store/cache
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required for cleanup behavior.

Done Definition:
- Expired cleanup is no longer capped at an arbitrary 1000 items.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Removed the hard-coded 1000-item cleanup cap.
- Added a regression test proving one cleanup pass removes more than 1000 expired entries while preserving live entries.
- Documented full-map in-process cleanup behavior.

Validation:
- `go test -timeout 20s ./store/cache`
- `go vet ./store/cache`
- `go run ./internal/checks/dependency-rules`
