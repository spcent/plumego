# Card 0775

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- docs/modules/x-data/README.md
Depends On:
- 0774-x-data-kvengine-snapshot-format-detection

Goal:
Make KVStore close timeout semantics preserve best-effort WAL flush and file close.

Scope:
- Ensure Close still attempts WAL flush and file close when worker shutdown exceeds CloseTimeout.
- Preserve ErrCloseTimeout visibility to callers.
- Add a regression test that acknowledged WAL data can recover after a close timeout.

Non-goals:
- Do not remove CloseTimeout.
- Do not change worker lifecycle outside shutdown ordering.
- Do not change WAL sync mode defaults.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Update x/data docs with close-timeout WAL behavior.

Done Definition:
- Close timeout returns ErrCloseTimeout while still attempting WAL flush/close.
- Recovery test proves acknowledged WAL data survives a timed-out close path.
- Tests and docs cover the behavior.
