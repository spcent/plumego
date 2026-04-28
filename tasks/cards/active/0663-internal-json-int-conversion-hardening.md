# Card 0663

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/jsonx
Owned Files: internal/jsonx/json.go, internal/jsonx/json_test.go, internal/pool/pool.go, internal/pool/pool_test.go
Depends On:

Goal:
Make internal JSON `int` extraction reject fractional and overflowing numbers consistently.

Scope:
- Add shared local int conversion helpers in `internal/jsonx` and `internal/pool`.
- Decode int extraction paths with `UseNumber` where needed so JSON numbers are checked before conversion.
- Apply the helper to top-level, nested, slice, map, and array-of-map int extraction functions.
- Add focused tests for fractional JSON numbers, fractional numeric strings, and `int` overflow.

Non-goals:
- Do not change `int64` extraction semantics.
- Do not make typed array/map extractors all-or-nothing for mixed element types in this card.
- Do not introduce cross-package dependencies between `jsonx` and `pool`.

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
Not required; this is internal correctness hardening.

Done Definition:
- Fractional and overflowing JSON integer extraction returns zero or skips the offending value instead of truncating/wrapping.
- Existing integer extraction behavior for valid integer values and integer strings remains intact.
- Focused and internal validation pass.

Outcome:
