# Card 1077

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
Depends On:
- tasks/cards/done/1065-sql-idempotency-duplicate-classifier.md

Goal:
Make SQL idempotency expired-key reclaim conditional on the row still being expired.

Scope:
- Replace unconditional Delete during PutIfAbsent reclaim with conditional expired-row delete.
- Preserve one-winner claim semantics under duplicate/reclaim races.
- Add regression coverage that a no-longer-expired row is not removed.

Non-goals:
- Do not introduce transactions or database-specific UPSERT syntax in this card.
- Do not change stable idempotency contracts.

Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go

Tests:
- go test -timeout 20s ./x/data/idempotency
- go test -timeout 20s ./store/idempotency

Docs Sync:
- None.

Done Definition:
- Reclaim cleanup only deletes rows that are still expired at the reclaim timestamp.
- Existing expired duplicate reclaim behavior still works.
- Targeted tests pass.

Outcome:
SQL PutIfAbsent reclaim now uses conditional expired-row deletion instead of unconditional Delete. deleteExpired reports RowsAffected, and tests cover that usable rows are not removed while expired rows are.

Validation:
- go test -timeout 20s ./x/data/idempotency
- go test -timeout 20s ./store/idempotency
