# Card 0729: Store Idempotency Conformance

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
- x/data/idempotency/idempotency_test.go
- x/data/idempotency/sql_test.go
Depends On:
- 0728

Goal:
Add conformance-style coverage for `store/idempotency.Store` so first-party implementations stay interchangeable.

Scope:
- Add a reusable test helper inside `store/idempotency` tests for store contract behavior.
- Apply the expected behavior to existing KV and SQL implementation tests.
- Cover missing, expired, duplicate, invalid-key, complete, delete, and response-copy expectations.

Non-goals:
- Do not add business-level hash conflict or replay policy.
- Do not move provider implementations into stable `store/idempotency`.

Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
- x/data/idempotency/idempotency_test.go
- x/data/idempotency/sql_test.go

Tests:
- go test -timeout 20s ./store/idempotency ./x/data/idempotency
- go test -race -timeout 60s ./store/idempotency ./x/data/idempotency
- go vet ./store/idempotency ./x/data/idempotency

Docs Sync:
- Required only if the conformance helper clarifies exported comments.

Done Definition:
- First-party KV and SQL idempotency implementations exercise the same stable behavior.
- Focused tests and vet pass.

Outcome:
