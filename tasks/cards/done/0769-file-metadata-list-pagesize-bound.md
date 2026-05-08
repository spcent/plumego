# Card 0769

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/file
Owned Files: x/data/file/metadata.go, x/data/file/metadata_test.go
Depends On:

Goal:

Add an explicit DB metadata List page-size upper bound so callers cannot request unbounded result sets.

Scope:

- Define a package-local maximum page size.
- Reject oversized PageSize values with a stable file error sentinel.
- Preserve the existing default for PageSize <= 0.
- Add focused tests for default and oversized page sizes.

Non-goals:

- Adding cursor-based pagination.
- Changing list ordering or filter semantics.
- Making the page-size limit configurable.

Files:

- x/data/file/metadata.go
- x/data/file/metadata_test.go

Tests:

- go test -race -timeout 60s ./x/data/file/...
- go test -timeout 20s ./x/data/file/...
- go vet ./x/data/file/...

Docs Sync:

- Not required; this is a defensive provider limit in an experimental adapter.

Done Definition:

- Oversized PageSize is rejected before SQL execution.
- Default PageSize behavior remains unchanged.
- Module tests and vet pass.

Outcome:

- DBMetadataManager.List now rejects PageSize values above a package-local maximum before issuing SQL.
- Oversized PageSize errors wrap store/file.ErrInvalidSize.
- Added fake-driver coverage proving oversized requests do not execute queries.
- Validation passed:
  - go test -race -timeout 60s ./x/data/file/...
  - go test -timeout 20s ./x/data/file/...
  - go vet ./x/data/file/...
