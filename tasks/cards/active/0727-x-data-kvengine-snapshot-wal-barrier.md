# Card 0727

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
Depends On:

Goal:
Make kvengine snapshot and WAL reset preserve acknowledged writes during concurrent snapshots.

Scope:
- Add a write/snapshot coordination path so Snapshot cannot omit acknowledged writes and then remove their WAL entries.
- Propagate WAL flush/reset errors instead of silently treating reset failures as success.
- Add a focused regression test where writes race with Snapshot and survive reopen.

Non-goals:
- Do not introduce a new WAL file format.
- Do not add transactions or batch APIs.
- Do not change serializer public types.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Not required unless the public durability contract text changes.

Done Definition:
- Snapshot has an explicit barrier against concurrent Set/Delete data loss.
- WAL reset/flush failures are returned to callers.
- kvengine normal and race tests pass.

Outcome:
- Added an operation mutex that serializes Set/Delete with Snapshot.
- Made WAL flush/reset failures return to Snapshot and Close instead of being ignored.
- Reset WAL by opening a fresh truncated writer only after flush/sync succeeds.
- Added a concurrent Snapshot regression test that verifies acknowledged writes survive reopen.

Validation:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine
