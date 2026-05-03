# Card 0717: Store Idempotency Contract Semantics

Priority: P1
State: active
Primary Module: store
Owned Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
- docs/modules/store/README.md

Goal:
Make the stable idempotency contract precise enough for interchangeable KV and SQL implementations.

Scope:
- Document `PutIfAbsent` behavior for absent, expired, in-progress, completed, and hash-conflict records.
- Document `Get` behavior for expired records and caller-owned response bytes.
- Document `Complete` behavior for missing or expired keys.
- Add tests that protect exported status and error contract wording where appropriate.

Non-goals:
- Do not add a concrete provider to stable `store/idempotency`.
- Do not define SQL schema, table naming, dialect policy, or duplicate-key mechanics.
- Do not add HTTP idempotency semantics.

Files:
- store/idempotency/store.go
- store/idempotency/store_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/idempotency ./x/data/idempotency
- go test -race -timeout 60s ./store/idempotency ./x/data/idempotency
- go vet ./store/idempotency ./x/data/idempotency

Docs Sync:
- Required for stable idempotency boundary wording.

Done Definition:
- Stable idempotency comments define implementation-neutral state behavior.
- Provider packages still satisfy the stable interface.
- Focused tests and vet pass.
