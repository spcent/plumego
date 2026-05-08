# Card 0125

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/file/types.go
- store/file/coverage_test.go
- store/idempotency/store.go
- store/idempotency/store_test.go
Depends On:
- 0124

Goal:
Add clone and validation helpers for mutable stable store contract types.

Scope:
- Add `File.Clone` and `PutOptions.CloneMetadata` or equivalent helpers for file metadata ownership.
- Add `Record.Clone`, `Status.Valid`, and record/key validation helpers for idempotency contracts.
- Add tests proving maps and slices are defensively copied by helpers.

Non-goals:
- Do not add concrete file or idempotency backends.
- Do not introduce tenant-aware metadata.
- Do not change existing interfaces.

Files:
- store/file/types.go
- store/file/coverage_test.go
- store/idempotency/store.go
- store/idempotency/store_test.go

Tests:
- go test -timeout 20s ./store/file ./store/idempotency
- go vet ./store/file ./store/idempotency

Docs Sync:
- Public comments only.

Done Definition:
- Stable contract callers and implementations have local helpers for defensive copying and validation.
- Existing interfaces remain compatible.
- Focused tests pass.

Outcome:
- Added `File.Clone` and `PutOptions.CloneMetadata` so file contract callers and implementations can make local defensive metadata copies.
- Added `Record.Clone`, `Status.Valid`, `ValidateKey`, `ValidateRecord`, and `ErrInvalidRecord` for idempotency contract validation and response copy ownership.
- Added tests proving file metadata maps, time pointers, and idempotency response slices are detached by the helpers.
- Refreshed `docs/stable-api/snapshots/store-head.snapshot`.

Validation:
- `go test -timeout 20s ./store/file ./store/idempotency`
- `go vet ./store/file ./store/idempotency`
- `go run ./internal/checks/extension-api-snapshot -module ./store/... -out docs/stable-api/snapshots/store-head.snapshot`
