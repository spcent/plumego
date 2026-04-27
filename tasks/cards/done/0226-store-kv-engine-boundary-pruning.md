# Card 0226

Priority: P1
State: done
Primary Module: store
Owned Files:
- `store/kv/kv.go`
- `store/kv/serializer.go`
- `store/kv/serializer_binary.go`
- `store/kv/serializer_json.go`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `store/idempotency`
- `security/jwt`
- `x/mq`
- `x/scheduler`
- `x/ai/distributed`
- `x/data`

Goal:
- Shrink stable `store/kv` to a small embedded KV primitive instead of a topology-heavy storage engine surface.
- Move WAL, snapshot/replay, serializer selection, and engine-tuning ownership out of the stable store root.

Problem:
- `store/kv/kv.go` currently exports a full embedded engine surface: `Shard`, `WALEntry`, `Snapshot`, serializer detection, compression toggles, shard count, background workers, and a metrics hook setter on the concrete store.
- `store/module.yaml` and `docs/modules/store/README.md` explicitly say topology-heavy data features and sharding defaults should stay out of stable `store`, but `store/kv` still owns exactly those concerns.
- `store/kv/serializer.go` exports engine-format plumbing (`SerializationFormat`, `Serializer`, `DetectFormat`, `DetectWALFormat`) that is persistence-engine internals rather than a stable application-facing KV primitive.
- Current consumers split across simple stable use cases (`store/idempotency`, `security/jwt`) and heavier durable-engine use cases (`x/mq`, `x/scheduler`, `x/ai/distributed`), which is a strong sign the stable KV root is carrying two ownership levels at once.

Scope:
- Define the minimum stable `store/kv` surface that should remain for embedded key-value persistence primitives.
- Move WAL, snapshot/replay, serializer/format detection, compression, shard-tuning, and engine-level metrics hook ownership into an extension-owned package under `x/data`.
- Update stable and extension consumers in the same change so simple stable callers depend only on the reduced KV primitive, while durability-heavy features import the extension-owned engine package.
- Remove engine-internal exported types and helpers from stable `store/kv` once the last caller is migrated.
- Sync store docs and manifest to the reduced KV boundary.

Non-goals:
- Do not redesign business-level idempotency, session, queue, or AI persistence semantics.
- Do not add tenant-aware storage policy to stable `store/kv`.
- Do not preserve removed WAL/serializer helpers through compatibility aliases.
- Do not move plain key-value CRUD semantics out of the stable store layer if they remain part of the final primitive surface.

Files:
- `store/kv/kv.go`
- `store/kv/serializer.go`
- `store/kv/serializer_binary.go`
- `store/kv/serializer_json.go`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `store/idempotency`
- `security/jwt`
- `x/mq`
- `x/scheduler`
- `x/ai/distributed`
- `x/data`

Tests:
- `go test -timeout 20s ./store/kv ./store/idempotency ./security/jwt ./x/mq ./x/scheduler ./x/ai/distributed/... ./x/data/...`
- `go test -race -timeout 60s ./store/kv ./store/idempotency ./security/jwt ./x/mq ./x/scheduler ./x/ai/distributed/... ./x/data/...`
- `go vet ./store/kv ./store/idempotency ./security/jwt ./x/mq ./x/scheduler ./x/ai/distributed/... ./x/data/...`

Docs Sync:
- Keep the store manifest and primer aligned on the rule that stable `store/kv` owns only the small embedded KV primitive, while WAL/snapshot/serializer/topology ownership lives in `x/data`.

Done Definition:
- Stable `store/kv` no longer exports WAL, serializer, snapshot/replay, sharding, compression, or engine-tuning surfaces that belong to the extension-owned KV engine.
- Simple stable consumers compile against the reduced KV primitive, while durability-heavy extension consumers compile against the new `x/data` owner package.
- Store docs and manifest describe the same reduced KV boundary the code implements.

Outcome:
- Completed.
- Reduced stable `store/kv` to a small file-backed KV primitive with TTL-aware CRUD, key scans, and basic stats.
- Moved the full durable engine surface to `x/data/kvengine`, including WAL, snapshots, serializer selection, compression, sharding, and engine-level tuning.
- Updated `x/mq` to use the new extension-owned engine while keeping stable consumers (`store/idempotency`, `security/jwt`, `x/scheduler`, `x/ai/distributed`) on the reduced stable primitive.
