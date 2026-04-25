# Card 2141: x/fileapi Handler Transport Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: x/fileapi
Owned Files:
- `x/fileapi/handler.go`
- `x/fileapi/handler_test.go`
- `x/fileapi/module.yaml`
- `docs/modules/x-fileapi/README.md`
Depends On:
- `2140-x-data-file-metadata-sql-shape-convergence.md`

Goal:
Converge `x/fileapi` handler transport logic so repeated tenant/file guards,
query parsing, and response payloads have one obvious local shape.

Problem:
`Handler` repeats tenant lookup, file ID extraction, metadata loading, and
tenant access checks across `Download`, `GetInfo`, `Delete`, and `GetURL`.
It also mixes typed storage results with ad hoc `map[string]string` and
`map[string]any` response payloads. `Download` writes headers and copies bytes
directly, ignoring copy errors after committing the status, while query parsing
for `expiry` silently falls back when the caller supplies an invalid value.

Scope:
- Add private helpers for tenant-required checks, file ID extraction, and
  metadata lookup with tenant authorization.
- Replace ad hoc success maps with small local response structs where the shape
  is part of the endpoint contract.
- Make invalid `expiry` values return a structured validation error instead of
  silently using the default.
- Add focused tests for repeated guard behavior and `Download` header behavior;
  keep streaming success as direct `net/http` output.

Non-goals:
- Do not change route paths.
- Do not change storage or metadata interfaces.
- Do not introduce a new response envelope.
- Do not add tenant resolution to stable middleware or stable store.

Files:
- `x/fileapi/handler.go`
- `x/fileapi/handler_test.go`
- `x/fileapi/module.yaml`
- `docs/modules/x-fileapi/README.md`

Tests:
- `go test -race -timeout 60s ./x/fileapi/...`
- `go test -timeout 20s ./x/fileapi/...`
- `go vet ./x/fileapi/...`

Docs Sync:
Update `docs/modules/x-fileapi/README.md` only if invalid query behavior or
response DTO names become documented endpoint behavior.

Done Definition:
- File handler tenant/file authorization has one private implementation path.
- Invalid query parameters fail through `contract.WriteError`.
- Success responses no longer use one-off maps where a local DTO is clearer.
- The listed validation commands pass.

Outcome:
- Added one private metadata/tenant authorization path shared by `Download`,
  `GetInfo`, `Delete`, and `GetURL`.
- Replaced local ad hoc success maps with endpoint DTO structs for delete, list,
  and temporary URL responses.
- Made invalid `expiry` query values return `contract.CodeInvalidQuery` with a
  structured `field` detail.
- Synced `docs/modules/x-fileapi/README.md` with the expanded malformed-query
  behavior.

Validation:
- `go test -race -timeout 60s ./x/fileapi/...`
- `go test -timeout 20s ./x/fileapi/...`
- `go vet ./x/fileapi/...`
