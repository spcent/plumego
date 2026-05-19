# Card 1338

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files: x/data/idempotency/sql.go, x/data/idempotency/sql_test.go
Depends On: 0773

Goal:

Make SQL idempotency expired duplicate reclaim use a transaction with conditional delete and insert, avoiding an observable delete+insert window.

Scope:

- On duplicate insert, reclaim with conditional DELETE and INSERT inside one transaction.
- Treat failed conditional delete as inserted=false,nil.
- Preserve RowsAffected errors and insert/commit errors.
- Add regression coverage that reclaim does not use check-then-delete outside the transaction path.

Non-goals:

- Adding database-specific UPSERT syntax.
- Changing Complete/Delete behavior.
- Changing KV idempotency behavior.

Files:

- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go

Tests:

- go test -race -timeout 60s ./x/data/idempotency
- go test -timeout 20s ./x/data/idempotency
- go vet ./x/data/idempotency

Docs Sync:

- Not required; this tightens provider correctness behind the existing contract.

Done Definition:

- Expired duplicate reclaim no longer does getRecord/deleteExpired/retry outside a transaction.
- Reclaim only succeeds when the expired row was conditionally deleted and the replacement insert commits.
- Module tests and vet pass.

Outcome:

- PutIfAbsent now handles duplicate insert by reclaiming expired rows inside a transaction.
- The reclaim path uses conditional DELETE followed by INSERT and COMMIT, without getRecord/check-then-delete.
- Added regression coverage that fails if reclaim selects before conditional delete.
- Validation passed:
  - go test -race -timeout 60s ./x/data/idempotency
  - go test -timeout 20s ./x/data/idempotency
  - go vet ./x/data/idempotency
