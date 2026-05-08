# Card 0759

Milestone:
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/kv.go
- x/data/idempotency/kv_test.go
- docs/modules/x-data/README.md
Depends On:
- 0758-x-data-rw-weighted-balancer-api

Goal:
Make the KV idempotency provider's single-process atomicity boundary explicit and testable.

Scope:
- Add exported documentation stating KV idempotency locking is in-process only.
- Keep SQL provider as the durable multi-process recommendation.
- Add focused tests that the shared in-process lock is used across wrappers sharing one KV instance.

Non-goals:
- Do not change the stable store/idempotency interface.
- Do not add CAS to store/kv in this card.
- Do not add distributed locking.

Files:
- x/data/idempotency/kv.go
- x/data/idempotency/kv_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/idempotency
- go test -race -timeout 60s ./x/data/idempotency
- go vet ./x/data/idempotency

Docs Sync:
- Document KV provider as local-process atomic and SQL provider as multi-process durable.

Done Definition:
- KV provider scope is explicit in code docs and module docs.
- Tests cover shared wrapper locking behavior.

Outcome:
- Added exported documentation that KV idempotency atomicity is process-local and shared only across wrappers over the same KV instance.
- Documented SQL provider as the multi-process durable coordination option.
- Added shared-wrapper Complete coverage.

Validation:
- `go test -timeout 20s ./x/data/idempotency`
- `go test -race -timeout 60s ./x/data/idempotency`
- `go vet ./x/data/idempotency`
