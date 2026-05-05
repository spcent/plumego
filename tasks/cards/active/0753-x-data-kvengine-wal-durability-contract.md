# Card 0753

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
- 0752-x-data-file-tenant-scoped-metadata-api

Goal:
Make kvengine WAL acknowledgement semantics explicit and durable enough for stable-readiness review.

Scope:
- Add an explicit synchronous WAL durability option.
- Ensure enabled sync writes flush and fsync WAL before memory state is updated.
- Keep async flushing available only as an explicit performance tradeoff.
- Harden WAL reset/snapshot directory sync where supported.

Non-goals:
- Do not replace the WAL format.
- Do not add external storage dependencies.
- Do not change stable store/kv.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Document sync WAL default and async durability tradeoff.

Done Definition:
- Set/Delete can be configured to return only after WAL fsync.
- WAL reset does not silently drop unflushed data.
- Tests cover the new option semantics.
