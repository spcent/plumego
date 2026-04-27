# Card 0327

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/db/sql.go
- store/db/sql_test.go
Depends On:
- 0326-store-cache-invalid-key-contract

Goal:
Expose transaction rollback failures instead of discarding them when the transaction function fails.

Scope:
- Preserve the original transaction function error.
- Join rollback errors into the returned `ErrTransactionFailed` chain.
- Add focused driver-backed coverage for rollback failure.

Non-goals:
- Do not change the `DB` interface.
- Do not add retry, timeout, health, metrics, or observability behavior.
- Do not change panic rollback behavior.

Files:
- store/db/sql.go
- store/db/sql_test.go

Tests:
- go test -timeout 20s ./store/db
- go test -race -timeout 60s ./store/db
- go vet ./store/db

Docs Sync:
- Not required.

Done Definition:
- Function errors and rollback errors are both observable through `errors.Is`.
- Existing transaction success, begin, nil, panic-adjacent behavior remains unchanged.
- DB targeted tests and vet pass.

Outcome:
- Joined rollback failures into the returned transaction error chain.
- Preserved the original transaction function error and `ErrTransactionFailed` sentinel.
- Added driver-backed coverage for rollback failure propagation.

Validation:
- go test -timeout 20s ./store/db
- go test -race -timeout 60s ./store/db
- go vet ./store/db
