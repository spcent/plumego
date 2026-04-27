# Card 0329

Milestone:
Recipe: specs/change-recipes/refine-docs.yaml
Priority: P2
State: active
Primary Module: store
Owned Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
Depends On:
- 0328-store-kv-value-too-large-contract

Goal:
Clarify `store/idempotency.Record` field ownership and stable duplicate-key semantics.

Scope:
- Document `RequestHash`, `Response`, and timestamp ownership expectations.
- State that implementations retaining response bytes must copy them.
- Add focused value-level coverage for response field representation.

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
- Record field comments explain caller/backend ownership without adding provider policy.
- Tests cover response byte representation without implying hidden copying by the value type.
- Idempotency targeted tests and vet pass.
