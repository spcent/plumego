# Card 0936

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
Depends On:
- 0734-x-data-sharding-sql-routing-semantics

Goal:
Tighten kvengine persistence and option semantics for stable readiness.

Scope:
- Make auto-detect explicitly configurable without bool zero-value ambiguity.
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

Outcome:
- Added an explicit disable path so callers can enforce the configured serializer.
- Made WAL replay return `ErrInvalidEntry` on decode errors or CRC mismatches.
- Included `ExpireAt` metadata in WAL CRC calculation.
- Removed stale transaction-support documentation and the unused transaction-aborted sentinel.
- Updated x/data docs for WAL corruption and auto-detect behavior.

Validation:
- GOCACHE=/private/tmp/plumego-go-build go test -timeout 20s ./x/data/kvengine
- GOCACHE=/private/tmp/plumego-go-build go test -race -timeout 60s ./x/data/kvengine
- GOCACHE=/private/tmp/plumego-go-build go vet ./x/data/kvengine
