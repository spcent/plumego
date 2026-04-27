# Card 0528

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: core
Owned Files:
- core/lifecycle.go
- core/lifecycle_test.go
Depends On: 2237

Goal:
Keep core connection tracking non-negative under duplicate or out-of-order terminal connection states.

Scope:
- Clamp connection tracker decrements at zero.
- Preserve normal `StateNew` / `StateClosed` / `StateHijacked` accounting.
- Add regression tests for duplicate terminal events.

Non-goals:
- Do not change public lifecycle APIs.
- Do not add readiness ownership to `core`.
- Do not change shutdown timeout configuration.

Files:
- `core/lifecycle.go`
- `core/lifecycle_test.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None expected; this hardens internal lifecycle accounting.

Done Definition:
- Active connection count never becomes negative.
- Existing lifecycle and server preparation tests continue to pass.

Outcome:
- Added a CAS-based decrement path that clamps active connection accounting at zero.
- Covered duplicate and out-of-order terminal connection states.
- Validation run: `go test -timeout 20s ./core/...`; `go vet ./core/...`.
