# Card 0731: SQL Idempotency Expired Reclaim

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
Depends On:

Goal:
Ensure SQL-backed idempotency keys can be reclaimed when an existing duplicate row is expired.

Scope:
- Make `PutIfAbsent` treat expired stored records as reclaimable.
- Keep usable in-progress or completed records blocking new claims.
- Preserve storage-agnostic stable `store/idempotency` semantics.
- Add tests for expired duplicate reclaim and usable duplicate blocking.

Non-goals:
- Do not change the stable `store/idempotency.Store` interface.
- Do not add migrations or table schema ownership to stable store.
- Do not add request hash conflict policy.

Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go

Tests:
- go test -timeout 20s ./x/data/idempotency ./store/idempotency
- go test -race -timeout 60s ./x/data/idempotency ./store/idempotency
- go vet ./x/data/idempotency ./store/idempotency

Docs Sync:
- Not required unless public behavior wording changes.

Done Definition:
- Expired duplicate rows no longer pin idempotency keys.
- Existing usable rows still return `inserted=false, nil`.
- Targeted tests, race tests, and vet pass.

Outcome:
- Added expired duplicate reclaim in `SQLStore.PutIfAbsent`: duplicate rows are
  inspected, expired records are deleted, and the insert is retried once.
- Kept usable duplicate rows blocking with `inserted=false, nil`.
- Added regression coverage for reclaiming an expired SQL duplicate and
  preserving the replacement record.

Validation:
- `go test -timeout 20s ./x/data/idempotency ./store/idempotency`
- `go test -race -timeout 60s ./x/data/idempotency ./store/idempotency`
- `go vet ./x/data/idempotency ./store/idempotency`
