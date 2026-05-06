# Card 0774

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/serializer.go
- x/data/kvengine/kv_test.go
- x/data/kvengine/serializer_test.go
- docs/modules/x-data/README.md
Depends On:
- 0773-x-data-file-path-and-cleanup-contract

Goal:
Make snapshot and WAL format detection fail closed and detect compressed snapshot payloads after decompression.

Scope:
- Return explicit errors for unknown snapshot and WAL magic instead of defaulting to binary.
- Detect compressed snapshot serialization format from the decompressed stream.
- Keep configured-format loading behavior available when auto-detect is disabled.
- Add regression coverage for compressed JSON snapshots and unknown snapshot/WAL content.

Non-goals:
- Do not add new serialization formats.
- Do not change snapshot filenames.
- Do not change WAL durability mode configuration.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/serializer.go
- x/data/kvengine/kv_test.go
- x/data/kvengine/serializer_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Update x/data docs with fail-closed format detection behavior.

Done Definition:
- Unknown snapshot/WAL formats return clear errors.
- Compressed JSON snapshots load correctly with auto-detect enabled.
- Tests and docs cover the behavior.
