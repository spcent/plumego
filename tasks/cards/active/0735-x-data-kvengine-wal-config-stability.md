# Card 0735

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
Depends On:
- 0734-x-data-sharding-sql-routing-semantics

Goal:
Tighten kvengine persistence and option semantics for stable readiness.

Scope:
- Make `AutoDetectFormat` explicitly configurable without bool zero-value ambiguity.
- Include expiry metadata in WAL integrity checks.
- Return WAL corruption errors instead of silently ignoring decode failures.
- Align comments/exported errors with implemented transaction support.

Non-goals:
- Do not introduce a transaction subsystem.
- Do not change snapshot file format unless required for WAL integrity.
- Do not redesign eviction.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Not required unless public option names or defaults change.

Done Definition:
- Explicitly disabling WAL format auto-detection works.
- Corrupted WAL entries fail replay visibly.
- WAL CRC covers expiry metadata.
- kvengine normal, race, and vet checks pass.
