# Card 0673

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/jsonx
Owned Files: internal/jsonx/json.go, internal/jsonx/json_test.go, internal/pool/pool.go, internal/pool/pool_test.go
Depends On:

Goal:
Prevent JSON float extraction helpers from returning non-finite values parsed from strings.

Scope:
- Make `jsonx` and `pool` float extraction helpers skip string values that parse to `NaN` or infinities.
- Preserve best-effort behavior for finite numbers and finite numeric strings.
- Add focused tests for scalar, slice, map, and array-map extraction.

Non-goals:
- Do not change integer extraction semantics.
- Do not change invalid JSON or missing-field behavior.

Files:
- internal/jsonx/json.go
- internal/jsonx/json_test.go
- internal/pool/pool.go
- internal/pool/pool_test.go

Tests:
- go test -timeout 20s ./internal/jsonx ./internal/pool
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; internal extraction hardening only.

Done Definition:
- Non-finite string float values are skipped or return the existing zero value.
- `jsonx` and `pool` remain aligned for equivalent helpers.

Outcome:
Done. `jsonx` and `pool` float extraction helpers now skip string values
that parse to `NaN` or infinities while preserving finite numeric strings.

Validation:
- go test -timeout 20s ./internal/jsonx ./internal/pool
- go test -timeout 20s ./internal/...
- go vet ./internal/...
