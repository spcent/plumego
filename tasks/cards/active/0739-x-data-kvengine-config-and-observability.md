# Card 0739

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
- 0738-x-data-idempotency-atomic-state-transitions

Goal:
Remove misleading kvengine defaults and make the exported metrics collector observe core operations.

Scope:
- Replace hidden relative DataDir defaults with explicit configuration errors for NewKVStore.
- Keep Default as the only convenience path that chooses a local directory.
- Wire metrics collector calls into Set, Get, and Delete paths.
- Add tests for missing DataDir and collector operation recording.

Non-goals:
- Do not remove exported option fields in this card.
- Do not change WAL serialization formats.
- Do not implement a WAL repair tool.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Update x/data docs for explicit DataDir requirements and metrics behavior.

Done Definition:
- NewKVStore without DataDir fails clearly instead of writing to cwd.
- Default remains documented as an explicit convenience constructor.
- Set/Get/Delete report metrics through MetricsObserver.

Outcome:
