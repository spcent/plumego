# Card 0722: Store Idempotency Deterministic Terminal Semantics

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
- x/data/idempotency/kv.go
- x/data/idempotency/idempotency_test.go
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
Depends On:
- 0721

Goal:
Make idempotency terminal operation semantics deterministic across stable contract and first-party implementations.

Scope:
- Define `Complete` on missing or expired records as `ErrNotFound`.
- Define `Delete` on missing records as `ErrNotFound`.
- Map lower-level KV missing/expired errors to stable idempotency sentinels.
- Update SQL implementation to avoid completing expired records.
- Add focused KV and SQL implementation tests.

Non-goals:
- Do not add provider-specific schema policy to `store/idempotency`.
- Do not add business-level hash conflict or replay decisions.
- Do not add new idempotency statuses.

Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
- x/data/idempotency/kv.go
- x/data/idempotency/idempotency_test.go
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go

Tests:
- go test -timeout 20s ./store/idempotency ./x/data/idempotency
- go test -race -timeout 60s ./store/idempotency ./x/data/idempotency
- go vet ./store/idempotency ./x/data/idempotency

Docs Sync:
- Required in `store/idempotency` package comments.

Done Definition:
- Stable idempotency comments no longer allow backend-dependent terminal operation results.
- KV and SQL idempotency implementations agree on missing and expired terminal operation behavior.
- Focused tests and vet pass.

Outcome:
- Tightened the stable `store/idempotency.Store` contract so `Complete` returns `ErrNotFound` for missing or expired records and `Delete` returns `ErrNotFound` for missing records.
- Mapped KV-backed missing and expired delete errors to stable idempotency `ErrNotFound`.
- Made SQL-backed `Complete` check usable record state before updating, so expired records are cleaned up and reported as not found.
- Made SQL-backed `Delete` inspect rows affected and return `ErrNotFound` for missing records.
- Added focused KV and SQL tests for missing delete and expired completion behavior.

Validation:
- go test -timeout 20s ./store/idempotency ./x/data/idempotency
- go test -race -timeout 60s ./store/idempotency ./x/data/idempotency
- go vet ./store/idempotency ./x/data/idempotency
