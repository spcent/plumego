# Card 1125

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md

Goal:
Give callers a stable way to classify corrupt KV state files and document persistence limits.

Scope:
- Add an `ErrCorruptState` sentinel for decode and invalid persisted key failures.
- Preserve lower-level wrapped errors for diagnostics.
- Add tests for corrupt JSON and invalid persisted keys.
- Sync store docs with corrupt-state and small-dataset durability guidance.

Non-goals:
- Do not add automatic repair.
- Do not add WAL, snapshots, or alternate serializers.
- Do not change state file format.

Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/kv
- go vet ./store/kv
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required for error and persistence semantics.

Done Definition:
- Corrupt state failures match `ErrCorruptState`.
- Existing startup validation remains fail-closed.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Added `kv.ErrCorruptState`.
- Wrapped state JSON decode failures and invalid persisted keys with `ErrCorruptState`.
- Preserved `ErrInvalidKey` in the invalid persisted key error chain.
- Documented corrupt-state classification and small-dataset persistence guidance.

Validation:
- `go test -timeout 20s ./store/kv`
- `go vet ./store/kv`
- `go run ./internal/checks/dependency-rules`
