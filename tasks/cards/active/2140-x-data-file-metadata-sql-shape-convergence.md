# Card 2140: x/data File Metadata SQL Shape Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data
Owned Files:
- `x/data/file/metadata.go`
- `x/data/file/metadata_test.go`
- `x/data/file/module.yaml`
- `docs/modules/x-data/README.md`
- `docs/modules/x-fileapi/README.md`
Depends On: none

Goal:
Make the database-backed file metadata manager explicit, testable, and less
duplicated without changing the app-facing file API.

Problem:
`DBMetadataManager` is documented as PostgreSQL-only and hardcodes `$N`
placeholders, but its constructor accepts a `*sql.DB` without validation and
returns the broad `MetadataManager` interface. `Get`, `GetByPath`, `GetByHash`,
and `List` duplicate the same select column list and scan/unmarshal block. The
manager also calls `time.Now()` directly in mutation paths, which makes expiry
and access-time behavior harder to test than other data modules that inject
clocks or config.

Scope:
- Add a small private shared select/scan helper for file metadata reads.
- Add explicit nil-db handling through an error-returning constructor or a
  documented sentinel path; preserve the existing constructor only if needed for
  compatibility.
- Add a narrow clock hook or private `now` function for `Delete` and
  `UpdateAccessTime` tests.
- Keep PostgreSQL-only behavior documented, or add a small dialect config only
  if the module manifest and docs are updated in the same card.

Non-goals:
- Do not redesign `datafile.Storage`.
- Do not change file API HTTP response shapes.
- Do not add a new SQL builder dependency.
- Do not widen stable `store/file`.

Files:
- `x/data/file/metadata.go`
- `x/data/file/metadata_test.go`
- `x/data/file/module.yaml`
- `docs/modules/x-data/README.md`
- `docs/modules/x-fileapi/README.md`

Tests:
- `go test -race -timeout 60s ./x/data/file`
- `go test -timeout 20s ./x/data/file`
- `go vet ./x/data/file`

Docs Sync:
Update module docs if constructor behavior, PostgreSQL-only behavior, or testable
clock behavior is made public.

Done Definition:
- File metadata scanning logic has one canonical implementation path.
- Nil database construction or use fails predictably and is covered.
- Delete/access-time timestamps are testable without sleeping or relying on wall
  clock precision.
- The listed validation commands pass.

Outcome:
