# Card 0319

Milestone:
Recipe: specs/change-recipes/refine-api.yaml
Priority: P2
State: active
Primary Module: store
Owned Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
Depends On:
- 0318-store-kv-expired-get-rollback

Goal:
Complete `store/idempotency` shared contract documentation and lock stable status values in tests.

Scope:
- Fix the package comment wiring example so the stable interface and concrete extension imports are unambiguous.
- Add missing exported type and method comments.
- Add tests for stable status wire values and record zero-value behavior.

Non-goals:
- Do not add a concrete idempotency implementation to stable `store`.
- Do not import `x/*` from production code.
- Do not change the `Store` interface.

Files:
- store/idempotency/store.go
- store/idempotency/store_test.go

Tests:
- go test -timeout 20s ./store/idempotency
- go test -race -timeout 60s ./store/idempotency
- go vet ./store/idempotency

Docs Sync:
- Not required; package comments only.

Done Definition:
- Public idempotency contract docs are complete and stable-layer scoped.
- Tests pin status values used by implementations.
- Production package remains interface-only and dependency-free.
