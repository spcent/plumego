# Card 0719: Store KV Stable Contract Convergence

Priority: P0
State: active
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md

Goal:
Converge `store/kv` on a stable contract that is honest about context, lifecycle, TTL, durability, and performance guarantees.

Scope:
- Audit all `store/kv` callers before changing exported method signatures.
- Decide whether to add context-aware operations or keep the primitive explicitly synchronous and caller-blocking.
- Align or document TTL, delete, exists, closed-store, and expired-key behavior.
- Clarify single-process, single-file durability limits and performance envelope.
- Add focused contract tests for whichever behavior is chosen.

Non-goals:
- Do not add WAL, snapshots, compression, sharding, serializers, or engine tuning to stable `store/kv`.
- Do not move `x/data/kvengine` behavior back into stable `store`.
- Do not silently break security/session/idempotency call sites.

Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/kv ./security/... ./x/data/idempotency ./x/tenant/session ./x/scheduler ./x/mq
- go test -race -timeout 60s ./store/kv ./security/... ./x/data/idempotency ./x/tenant/session ./x/scheduler ./x/mq
- go vet ./store/kv ./security/... ./x/data/idempotency ./x/tenant/session ./x/scheduler ./x/mq

Docs Sync:
- Required for stable KV boundary wording.

Done Definition:
- KV contract is explicit about context and durability limits.
- Existing direct callers are updated or intentionally left on the documented API.
- Focused tests and vet pass.
