# Card 0664

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: internal/jsonx
Owned Files: internal/jsonx/json.go, internal/jsonx/json_test.go, internal/pool/pool.go, internal/pool/pool_test.go
Depends On: 0663

Goal:
Decide and enforce whether typed JSON collection extractors are all-or-nothing or partial best-effort.

Scope:
- Audit doc comments and tests for string, bool, numeric, map, and array-of-map collection extractors.
- Either update implementation to return nil on mixed invalid elements, or update comments/tests to document partial best-effort behavior.
- Keep `jsonx` and `pool` behavior aligned for equivalent helpers.

Non-goals:
- Do not change scalar field extraction.
- Do not introduce new public helper families.

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
Not required unless behavior is intentionally documented in repository docs later.

Done Definition:
- Typed collection extractor behavior is internally consistent and test-covered.
- Equivalent `jsonx` and `pool` helpers do not disagree silently.

Outcome:
