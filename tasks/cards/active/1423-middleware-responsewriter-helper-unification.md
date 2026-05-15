# Card 1423

Milestone: v1-breaking-normalization
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
- middleware/*
- middleware/internal/transport/*
Depends On:
- 1421

Goal:
- Consolidate duplicate response writer recording helpers while preserving
  transport-only middleware behavior.

Scope:
- Enumerate status/bytes recorder implementations across middleware packages.
- Keep one internal helper that preserves optional `http.ResponseWriter`
  interfaces where required.
- Migrate middleware packages to the canonical helper.
- Keep error responses on `contract.WriteError` when JSON transport errors are
  emitted.
- Add or preserve conformance tests for ordering, panic, timeout, cancellation,
  and wrapped writer behavior.

Non-goals:
- Do not inject business services or domain DTOs.
- Do not move middleware into `core` or `router`.
- Do not change middleware signature away from `func(http.Handler) http.Handler`.

Files:
- middleware/internal/transport/http.go
- middleware/internal/transport/response_buffer.go
- middleware/*/*.go
- middleware/conformance/*_test.go

Tests:
- go test -timeout 20s ./middleware/...
- go vet ./middleware/...
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update middleware docs only if helper behavior or public construction changes.

Done Definition:
- Duplicate recorder helpers are removed or delegated to the canonical helper.
- Middleware conformance tests pass.
- Middleware remains transport-only.

Outcome:

