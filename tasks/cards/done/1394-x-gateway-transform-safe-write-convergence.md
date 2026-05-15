# Card 1394

Milestone: v1-cleanup-phase-2
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/gateway/transform
Owned Files:
- x/gateway/transform/transform.go
- x/gateway/transform/transform_test.go
Depends On:

Goal:
- Remove the duplicate local HTTP safe-write helper from `x/gateway/transform` without changing response transform behavior.

Scope:
- Replace the package-local `safeWrite` helper with `internal/httputil.SafeWrite`.
- Preserve nil-writer no-op behavior and `X-Content-Type-Options: nosniff` semantics.
- Keep the existing buffering response recorder local to `x/gateway/transform` in this card.
- Add or keep focused tests proving transformed responses still write the transformed body and headers.

Non-goals:
- Do not replace the transform recorder with `internal/httputil.ResponseRecorder`; that helper writes through to the wrapped writer and would change transform semantics.
- Do not change public gateway APIs.
- Do not refactor cache, proxy, or unrelated gateway packages.

Files:
- x/gateway/transform/transform.go
- x/gateway/transform/transform_test.go

Tests:
- go test -timeout 20s ./x/gateway/transform
- go vet ./x/gateway/transform
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless behavior or public API text changes during execution.

Done Definition:
- `x/gateway/transform` no longer defines a duplicate safe-write helper.
- Response transformers still buffer upstream responses before final writes.
- Focused transform tests and dependency checks pass.

Outcome:
- Replaced the package-local `safeWrite` helper with `internal/httputil.SafeWrite`.
- Kept the transform response recorder local so response transformers still buffer upstream output before the final write.
- Added coverage that transformed responses preserve status, write the transformed body, and apply `X-Content-Type-Options: nosniff`.
- Validation:
  - `go test -timeout 20s ./x/gateway/transform`
  - `go vet ./x/gateway/transform`
  - `go run ./internal/checks/dependency-rules`
