# Card 1303

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/file
Owned Files: x/data/file/metadata.go, x/data/file/metadata_test.go
Depends On:

Goal:

Make DBMetadataManager.Delete surface RowsAffected driver errors instead of mapping them to ErrNotFound.

Scope:

- Check RowsAffected errors from Exec results.
- Preserve ErrNotFound only when RowsAffected succeeds and returns zero.
- Add regression coverage with a fake driver result.

Non-goals:

- Changing metadata Delete SQL.
- Changing other metadata manager methods.
- Adding database-driver-specific handling.

Files:

- x/data/file/metadata.go
- x/data/file/metadata_test.go

Tests:

- go test -race -timeout 60s ./x/data/file/...
- go test -timeout 20s ./x/data/file/...
- go vet ./x/data/file/...

Docs Sync:

- Not required; this fixes error preservation.

Done Definition:

- RowsAffected errors are returned to callers.
- Zero affected rows still map to file.ErrNotFound.
- Module tests and vet pass.

Outcome:

- DBMetadataManager.Delete now returns RowsAffected driver errors directly.
- Zero affected rows still map to file.ErrNotFound.
- Added fake-driver coverage for RowsAffected failure.
- Validation passed:
  - go test -race -timeout 60s ./x/data/file/...
  - go test -timeout 20s ./x/data/file/...
  - go vet ./x/data/file/...
