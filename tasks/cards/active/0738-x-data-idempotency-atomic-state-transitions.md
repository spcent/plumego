# Card 0738

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/kv.go
- x/data/idempotency/sql.go
- x/data/idempotency/kv_test.go
- x/data/idempotency/sql_test.go
- docs/modules/x-data/README.md
Depends On:
- 0737-x-data-sharding-routing-and-metrics-stability

Goal:
Make idempotency PutIfAbsent and Complete behavior deterministic under concurrency and stale records.

Scope:
- Guard KVStore Complete with the same store-level critical section used by PutIfAbsent.
- Make SQL Complete transition only active in-progress, non-expired records.
- Preserve existing public interfaces while tightening semantics.
- Add tests for expired records, already-completed records, and concurrent wrapper behavior.

Non-goals:
- Do not change the exported Store interface.
- Do not introduce a new stable store CAS API in this card.
- Do not add database-driver-specific duplicate error adapters beyond current supported dialects.

Files:
- x/data/idempotency/kv.go
- x/data/idempotency/sql.go
- x/data/idempotency/kv_test.go
- x/data/idempotency/sql_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/idempotency
- go test -race -timeout 60s ./x/data/idempotency
- go vet ./x/data/idempotency

Docs Sync:
- Update x/data docs for Complete transition requirements.

Done Definition:
- Complete cannot complete missing, expired, or already-final records.
- KV wrapper methods do not race internally across PutIfAbsent, Complete, and Delete.
- SQL Complete uses a conditional update instead of a key-only update.

Outcome:
