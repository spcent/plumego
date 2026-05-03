# Card 0721: Store KV Constructor And Package Contract

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P0
State: active
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md
Depends On:
- 0720

Goal:
Converge `store/kv` on explicit constructor and naming behavior before v1 freezes the public surface.

Scope:
- Rename the package clause from `kvstore` to `kv` while preserving import-path compatibility for existing explicit aliases.
- Add a stable `ErrInvalidConfig` sentinel for option validation.
- Require an explicit `Options.DataDir` instead of silently writing to `./data`.
- Keep existing capacity defaults for callers that provide a data directory.
- Update tests and store module documentation.

Non-goals:
- Do not add WAL, snapshots, cross-process locks, serializers, or durability tuning.
- Do not change the `KVStore` method set.
- Do not migrate downstream explicit import aliases unless required by compilation.

Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/kv ./security/... ./x/data/idempotency ./x/tenant/session ./x/scheduler ./x/mq ./x/ai/distributed
- go test -race -timeout 60s ./store/kv ./security/... ./x/data/idempotency ./x/tenant/session ./x/scheduler ./x/mq ./x/ai/distributed
- go vet ./store/kv ./security/... ./x/data/idempotency ./x/tenant/session ./x/scheduler ./x/mq ./x/ai/distributed

Docs Sync:
- Required for `DataDir`, package naming, and constructor error semantics.

Done Definition:
- `store/kv` package name aligns with its import path.
- Empty or whitespace-only `DataDir` returns an `ErrInvalidConfig`-wrapped error.
- Focused downstream tests and vet pass.

Outcome:
