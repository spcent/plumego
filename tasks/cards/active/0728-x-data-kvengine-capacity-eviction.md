# Card 0728

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
Depends On:
- 0727-x-data-kvengine-snapshot-wal-barrier

Goal:
Prevent kvengine Set from spinning forever when global capacity pressure is caused by other shards.

Scope:
- Replace target-shard-only eviction loops with progress-checked eviction across shards or explicit value-too-large rejection.
- Preserve existing LRU behavior as much as possible while guaranteeing loop progress.
- Add focused tests for max-entry and max-memory pressure when the target shard has no evictable tail.

Non-goals:
- Do not redesign the full cache admission policy.
- Do not add background compaction.
- Do not change the public Options shape unless necessary.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Not required unless a new public error is introduced.

Done Definition:
- Set always terminates under max-entry and max-memory pressure.
- Oversized single writes fail clearly instead of spinning.
- kvengine normal and race tests pass.

Outcome:
