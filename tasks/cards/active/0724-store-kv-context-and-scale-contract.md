# Card 0724

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md
- docs/stable-api/snapshots/store-head.snapshot
Depends On: 0723

Goal:
Clarify and enforce the stable KV context and small-dataset persistence contract.

Scope:
- Re-check caller context after acquiring the KV store lock and before durable persistence starts.
- Document that in-flight filesystem persistence is not interruptible once started.
- Document the small-dataset cost model for full-state JSON persistence.
- Add focused tests for canceled context behavior around write entrypoints.

Non-goals:
- Do not add WAL, snapshots, serializer selection, compression, or shard tuning.
- Do not change the stable KV file format.
- Do not add provider-specific durable engine behavior.

Files:
- `store/kv/kv.go`
- `store/kv/kv_test.go`
- `docs/modules/store/README.md`
- `docs/stable-api/snapshots/store-head.snapshot`

Tests:
- `go test -race -timeout 60s ./store/kv`
- `go test -timeout 20s ./store/...`
- `go vet ./store/...`

Docs Sync:
- Required for KV context and persistence scale semantics.

Done Definition:
- Context-aware write methods reject canceled contexts after lock acquisition and before persistence.
- Docs state the persistence cost model and non-interruptible filesystem phase.
- Targeted store tests, vet, and snapshot are updated.

Outcome:
- Pending.
