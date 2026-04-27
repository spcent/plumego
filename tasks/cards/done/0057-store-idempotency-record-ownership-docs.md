# Card 0057

Milestone:
Recipe: specs/change-recipes/refine-docs.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
Depends On:
- 0056-store-kv-value-too-large-contract

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

Outcome:
- Documented idempotency record key, request hash, response, and timestamp ownership.
- Clarified that retaining implementations must defensively copy response bytes.
- Extended record field tests for request hash and caller-owned response bytes.

Validation:
- go test -timeout 20s ./store/idempotency
- go test -race -timeout 60s ./store/idempotency
- go vet ./store/idempotency
