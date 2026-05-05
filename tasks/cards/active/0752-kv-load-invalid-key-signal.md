# Card 0752

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P2
State: active
Primary Module: store/kv
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:

Goal:
Stop silently dropping invalid persisted keys during KV state load.

Scope:
- Return a detectable error when persisted state contains an invalid key.
- Preserve normal load for valid persisted state.
- Add regression coverage with a hand-written invalid state file.

Non-goals:
- Do not add recovery modes, WAL, snapshots, or migration tooling.
- Do not change the public KV API shape.

Files:
- store/kv/kv.go
- store/kv/kv_test.go

Tests:
- go test -timeout 20s ./store/kv
- go test -timeout 20s ./store/...

Docs Sync:
- docs/modules/store/README.md if load behavior documentation needs adjustment.

Done Definition:
- Invalid persisted keys fail load with ErrInvalidKey in the error chain.
- Valid persisted states still load.
- Targeted tests pass.

Outcome:

