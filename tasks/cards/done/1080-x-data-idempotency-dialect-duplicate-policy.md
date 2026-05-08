# Card 1080

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
- x/data/idempotency/module.yaml
- docs/modules/x-data/README.md
Depends On:
- 0746-x-data-kvengine-api-lifecycle-freeze

Goal:
Make SQL duplicate-key handling explicit instead of relying only on broad string matching.

Scope:
- Add an optional DuplicateError function to SQLConfig.
- Use the configured classifier before falling back to conservative built-in matching.
- Document the fallback as compatibility behavior, not the stable dialect contract.
- Add tests for custom duplicate classifier behavior.

Non-goals:
- Do not import database driver packages.
- Do not change the store/idempotency stable interface.
- Do not implement distributed KV CAS in this card.

Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
- x/data/idempotency/module.yaml
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/idempotency
- go test -race -timeout 60s ./x/data/idempotency
- go vet ./x/data/idempotency

Docs Sync:
- Update x/data docs for duplicate error classifier and KV same-process atomicity boundary.

Done Definition:
- SQLStore can be configured with an explicit duplicate classifier.
- Existing fallback behavior remains covered by tests.
- Docs distinguish configured duplicate handling from fallback matching.

Outcome:
- Added `SQLConfig.DuplicateError` as an explicit duplicate-key classifier hook.
- Kept the existing broad duplicate string matcher as compatibility fallback.
- Added tests for custom classifier matching, fallback matching, and `sql.ErrNoRows`
  exclusion.
- Updated the idempotency manifest and x/data docs.

Validation:
- `go test -timeout 20s ./x/data/idempotency`
- `go test -race -timeout 60s ./x/data/idempotency`
- `go vet ./x/data/idempotency`
