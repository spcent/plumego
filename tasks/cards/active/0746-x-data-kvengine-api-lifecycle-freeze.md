# Card 0746

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- x/data/kvengine/module.yaml
- docs/modules/x-data/README.md
Depends On:
- 0745-x-data-sharding-cross-shard-first-success

Goal:
Make kvengine lifecycle and metrics collector behavior safe enough to document before API freeze.

Scope:
- Make Close idempotent and return the original close result on repeated calls.
- Protect metrics collector get/set/use with synchronization.
- Update module manifest to include current public convenience and metrics entrypoints.
- Add tests for repeated Close and concurrent metrics collector access.

Non-goals:
- Do not remove AutoDetectFormat or DisableAutoDetect in this card.
- Do not change WAL or serializer formats.
- Do not mark kvengine stable.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- x/data/kvengine/module.yaml
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Update x/data docs for Close idempotency, metrics collector concurrency, and remaining option-shape blocker.

Done Definition:
- Repeated Close is safe and deterministic.
- Metrics collector access is race-safe.
- Manifest reflects exported kvengine entrypoints used by docs/tests.

Outcome:
