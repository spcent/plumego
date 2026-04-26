# Card 0323

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:
- 0322-store-db-query-nil-rows-guard

Goal:
Apply the KV key contract consistently when loading persisted state.

Scope:
- Reuse the KV key validator while loading `store.json`.
- Skip invalid persisted keys instead of reintroducing keys runtime operations reject.
- Add regression coverage for a persisted empty key.

Non-goals:
- Do not change the on-disk JSON shape.
- Do not add WAL, snapshot, serializer, shard, or tenant behavior.
- Do not change runtime key validation semantics.

Files:
- store/kv/kv.go
- store/kv/kv_test.go

Tests:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv

Docs Sync:
- Not required.

Done Definition:
- Invalid persisted keys do not appear in memory after `NewKVStore`.
- Valid persisted keys still load with recomputed sizes.
- KV targeted tests and vet pass.

Outcome:
- Reused KV key validation while loading persisted state.
- Skipped invalid persisted keys and let startup normalization persist the cleaned state.
- Added regression coverage for persisted empty keys.

Validation:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv
