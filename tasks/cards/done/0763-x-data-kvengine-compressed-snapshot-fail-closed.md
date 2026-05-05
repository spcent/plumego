# Card 0763

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- docs/modules/x-data/README.md
Depends On:
- 0762-x-data-kvengine-autodetect-option-freeze

Goal:
Make compressed snapshot loading fail closed when gzip initialization fails.

Scope:
- Return a clear error when `EnableCompression` is set and the snapshot is not valid gzip data.
- Add a regression test for invalid compressed snapshot input.
- Keep uncompressed snapshot loading behavior unchanged.

Non-goals:
- Do not change snapshot file naming.
- Do not add a new compression algorithm.
- Do not change WAL replay behavior.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Update x/data docs to record compressed snapshot fail-closed behavior.

Done Definition:
- Invalid compressed snapshots return an error.
- Uncompressed snapshots still load under the default configuration.
- Tests and docs cover the behavior.

Outcome:
- Changed compressed snapshot loading to return an error when gzip reader
  initialization fails.
- Added a regression test for invalid compressed snapshot data.
- Updated x/data docs with the fail-closed compressed snapshot contract.

Validation:
- `go test -timeout 20s ./x/data/kvengine`
- `go test -race -timeout 60s ./x/data/kvengine`
- `go vet ./x/data/kvengine`
