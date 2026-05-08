# Card 0943: KV Performance Envelope

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md
Depends On:
- 0735

Goal:
Make the stable KV full-snapshot write model explicit, guarded, and measurable.

Scope:
- Add tests that lock in small embedded KV defaults and large-dataset escape guidance.
- Add benchmark or test evidence for overwrite/delete snapshot behavior.
- Improve docs so stable users understand the O(N) mutation model and when to use `x/data/kvengine`.

Non-goals:
- Do not add WAL, snapshots, serializers, compression, or sharding to stable `store/kv`.
- Do not change the public `KVStore` method set.
- Do not import `x/*` from stable store.

Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv

Docs Sync:
- Required for KV performance envelope.

Done Definition:
- KV stable docs explicitly describe full-snapshot O(N) mutation behavior.
- Tests protect intended defaults and small-store scope.
- Targeted tests, race tests, and vet pass.

Outcome:
- Documented that `store/kv` mutations rewrite the full normalized JSON state
  and are O(N) in retained entries.
- Added a default-envelope regression test that locks the embedded KV defaults
  to 4096 entries and 32 MiB.
- Kept durable engine concerns such as WAL, snapshots, serializers, compression,
  and sharding out of stable `store/kv`.

Validation:
- `go test -timeout 20s ./store/kv`
- `go test -race -timeout 60s ./store/kv`
- `go vet ./store/kv`
