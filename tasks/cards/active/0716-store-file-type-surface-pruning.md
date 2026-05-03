# Card 0716: Store File Type Surface Pruning

Priority: P1
State: active
Primary Module: store
Owned Files:
- store/file/types.go
- store/file/coverage_test.go
- store/file/README.md
- docs/modules/store/README.md

Goal:
Prune unused backend-shaped `store/file` type surface so the stable package stays focused on transport-agnostic storage contracts.

Scope:
- Remove unused stable `Query` type if no non-test callers remain.
- Remove SQL-oriented struct tags from stable file metadata types.
- Keep `Storage`, `File`, `PutOptions`, and `FileStat` transport-agnostic.
- Update docs that describe the stable file type surface.

Non-goals:
- Do not change `x/data/file` metadata query behavior.
- Do not add file backend/provider configuration.
- Do not add helper APIs back into `store/file`.

Files:
- store/file/types.go
- store/file/coverage_test.go
- store/file/README.md
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/file ./x/data/file ./x/fileapi
- go test -race -timeout 60s ./store/file ./x/data/file ./x/fileapi
- go vet ./store/file ./x/data/file ./x/fileapi

Docs Sync:
- Required for stable file boundary wording.

Done Definition:
- `store/file` no longer exposes unused query/provider-shaped surface.
- Downstream file packages compile and pass focused tests.
- Docs match the reduced stable file contract.
