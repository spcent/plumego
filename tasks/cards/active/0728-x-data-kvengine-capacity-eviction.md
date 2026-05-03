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
- Added explicit single-value memory admission before WAL writes.
- Replaced target-shard-only capacity loops with progress-checked cross-shard LRU eviction.
- Preserved existing entries during replacement by skipping the target key while making room.
- Added max-entry and max-memory regression tests where the new key's shard starts empty.

Validation:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine
