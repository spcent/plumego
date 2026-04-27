# Card 0043

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/db/sql.go
- store/db/sql_test.go
Depends On:
- 0042-store-kv-load-normalization

Goal:
Harden `store/db` helper resource edges without adding DB policy ownership.

Scope:
- Treat an injected opener returning `nil, nil` as an invalid connection failure.
- Treat `BeginTx` returning `nil, nil` as a transaction failure.
- Propagate `Rows.Close` errors from `QueryRowStrict` and `ScanRows` while preserving existing sentinel matching.
- Add focused tests for nil resources and close-error propagation.

Non-goals:
- Do not add retry, timeout, health, metrics, analytics, or pool-stat behavior.
- Do not change the `DB` interface.
- Do not introduce non-stdlib dependencies.

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
- Nil DB/transaction resources from injected code fail closed.
- Rows close failures are visible to callers.
- Existing SQL helper tests and context-ownership behavior still pass.

Outcome:
- Added fail-closed handling for injected openers returning `nil, nil`.
- Added fail-closed handling for `BeginTx` returning `nil, nil`.
- Propagated rows close errors from `QueryRowStrict` and `ScanRows` with `errors.Join` while keeping `ErrQueryFailed` matchable.
- Added targeted tests for nil DB resources, nil transactions, and rows close failures.

Validation:
- go test -timeout 20s ./store/db
- go test -race -timeout 60s ./store/db
- go vet ./store/db
