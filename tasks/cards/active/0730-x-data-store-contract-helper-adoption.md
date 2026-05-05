# Card 0730

Milestone:
Recipe: specs/change-recipes/refactor.yaml
Priority: P1
State: active
Primary Module: x/data
Owned Files:
- x/data/idempotency/store.go
- x/data/idempotency/kv.go
- x/data/idempotency/sql.go
- x/data/file/local.go
- x/data/file/s3.go
Depends On:
- 0729

Goal:
Adopt stable store clone and validation helpers in extension implementations so stable contracts and provider behavior stay aligned.

Scope:
- Re-export `ErrInvalidRecord`, `ValidateKey`, and `ValidateRecord` from `x/data/idempotency`.
- Use stable idempotency validation helpers in KV and SQL providers.
- Defensively copy idempotency responses and file metadata when providers retain or return mutable contract data.

Non-goals:
- Do not change file backend provider features.
- Do not add SQL schema management.
- Do not change tenant-aware file API contracts.

Files:
- x/data/idempotency/store.go
- x/data/idempotency/kv.go
- x/data/idempotency/sql.go
- x/data/file/local.go
- x/data/file/s3.go

Tests:
- go test -timeout 20s ./x/data/idempotency ./x/data/file
- go vet ./x/data/idempotency ./x/data/file

Docs Sync:
- None unless behavior comments need clarification.

Done Definition:
- Providers use stable validation helpers instead of local drift-prone key checks.
- Stored/replayed response bytes are defensively copied.
- File metadata returned by local/S3 providers no longer aliases caller metadata maps.

Outcome:

Validation:
