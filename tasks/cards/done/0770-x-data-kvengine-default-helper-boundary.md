# Card 0770

Milestone:
Recipe: specs/change-recipes/refactor-api.yaml
Priority: P3
State: done
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- x/data/kvengine/module.yaml
- docs/modules/x-data/README.md
Depends On:
- 0769-x-data-rw-sql-lock-routing-precision

Goal:
Remove cwd-writing surprises from the exported kvengine default helper surface.

Scope:
- Replace or constrain `Default()` so production callers must choose an explicit data directory.
- Update tests and manifest for the chosen public helper shape.
- Document any compatibility break while x/data remains experimental.

Non-goals:
- Do not change WAL or snapshot formats.
- Do not change serializer defaults.
- Do not move the helper into application wiring.

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
- Update x/data docs with the explicit-path helper contract.

Done Definition:
- Exported kvengine helpers no longer silently write to relative `data`.
- Tests and manifest reflect the public API.
- Docs identify the stable helper path.

Outcome:
- Changed `Default` to require a caller-provided data directory.
- Updated tests to use temporary directories and reject an empty default path.
- Updated x/data docs with the explicit-path helper contract.

Validation:
- `go test -timeout 20s ./x/data/kvengine`
- `go test -race -timeout 60s ./x/data/kvengine`
- `go vet ./x/data/kvengine`
