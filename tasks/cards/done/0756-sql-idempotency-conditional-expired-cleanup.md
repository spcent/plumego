# Card 0756

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
Depends On:

Goal:
Make SQL idempotency Get clean up expired records with a conditional delete instead of deleting by key unconditionally.

Scope:
- Replace Get's unconditional Delete call for expired records with deleteExpired.
- Use one captured now value for expiry decision and cleanup.
- Add or update tests covering expired Get cleanup without broadening SQL API.

Non-goals:
- Do not redesign SQL upsert or duplicate classifier behavior.
- Do not add driver-specific dependencies.

Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go

Tests:
- go test ./x/data/idempotency

Docs Sync:
- Not required unless public behavior or docs change.

Done Definition:
- Get treats expired records as absent and cleanup is protected by an expires_at condition.
- SQL idempotency tests pass.

Outcome:
- Updated SQLStore.Get to use a captured now value and deleteExpired for expired
  cleanup instead of unconditional Delete.
- Added a regression test that rejects unconditional DELETE for the expired Get
  path and verifies conditional cleanup removes the row.
- Validated with:
  - go test -timeout 20s ./x/data/idempotency
  - go test -race -timeout 60s ./x/data/idempotency
  - go vet ./x/data/idempotency
