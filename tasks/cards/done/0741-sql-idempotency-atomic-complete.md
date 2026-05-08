# Card 0741

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
Depends On:

Goal:
Make SQL idempotency Complete enforce expiry/status state in the UPDATE itself.

Scope:
- Replace check-then-update with a conditional UPDATE.
- Treat expired or missing records as ErrNotFound.
- Keep response byte ownership and existing config validation behavior.
- Add tests for expired records and changed rows.

Non-goals:
- Do not change the stable idempotency interface.
- Do not add database-driver-specific dependencies.

Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go

Tests:
- go test -timeout 20s ./x/data/idempotency
- go test -timeout 20s ./store/idempotency

Docs Sync:
- None unless documented SQL provider semantics need clarification.

Done Definition:
- Complete cannot complete expired records.
- RowsAffected drives ErrNotFound for non-usable rows.
- Targeted tests pass.

Outcome:
SQL Complete now uses a conditional UPDATE that only completes records whose expires_at is null or still in the future. RowsAffected=0 returns ErrNotFound and triggers conditional expired-row cleanup. Added regression coverage that expired Complete removes the stale row.

Validation:
- go test -timeout 20s ./x/data/idempotency
- go test -timeout 20s ./store/idempotency
