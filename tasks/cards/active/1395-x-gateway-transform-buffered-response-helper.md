# Card 1395

Milestone: v1-cleanup-phase-2
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/httputil
Owned Files:
- internal/httputil/buffered_response.go
- internal/httputil/buffered_response_test.go
- x/gateway/transform/transform.go
- x/gateway/transform/transform_test.go
Depends On:
- 1394

Goal:
- Extract the non-forwarding buffered response recorder used by `x/gateway/transform` into a shared internal helper.

Scope:
- Add an `internal/httputil` helper that records status, headers, and body without writing through to the wrapped writer.
- Switch `x/gateway/transform` to the new buffered helper.
- Preserve status code defaults, header copy behavior, body replacement, and final transformed write behavior.
- Cover the helper with focused tests that distinguish it from the write-through `ResponseRecorder`.

Non-goals:
- Do not change `internal/httputil.ResponseRecorder`; existing callers may rely on write-through behavior.
- Do not add public API outside `internal/httputil`.
- Do not change gateway transformer interfaces or external package contracts.

Files:
- internal/httputil/buffered_response.go
- internal/httputil/buffered_response_test.go
- x/gateway/transform/transform.go
- x/gateway/transform/transform_test.go

Tests:
- go test -timeout 20s ./internal/httputil ./x/gateway/transform
- go vet ./internal/httputil ./x/gateway/transform
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless the helper changes documented gateway behavior.

Done Definition:
- `x/gateway/transform` no longer owns recorder/write helper duplication.
- The new helper explicitly documents and tests non-forwarding semantics.
- Gateway transform behavior remains externally unchanged.

Outcome:

