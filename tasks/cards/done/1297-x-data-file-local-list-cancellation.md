# Card 1297

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/local.go
- x/data/file/local_test.go
- docs/modules/x-data/README.md
Depends On:
- 0766-x-data-file-metadata-list-tenant-contract

Goal:
Make local file listing cancellation-aware and define missing-prefix behavior.

Scope:
- Stop filesystem traversal promptly when context is canceled.
- Return an empty result for a missing prefix directory.
- Preserve tenant path isolation and existing pagination limit behavior.

Non-goals:
- Do not change upload/download behavior.
- Do not add filesystem watchers.
- Do not change metadata storage.

Files:
- x/data/file/local.go
- x/data/file/local_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Update x/data docs with local list cancellation and missing-prefix semantics.

Done Definition:
- Canceled context aborts listing with context error.
- Missing prefixes return an empty list.
- Tests and docs cover the behavior.

Outcome:
- Made `LocalStorage.List` check context before and during traversal.
- Defined missing-prefix listing as an empty result.
- Switched listing to `filepath.WalkDir` and preserved limit behavior.

Validation:
- `go test -timeout 20s ./x/data/file`
- `go test -race -timeout 60s ./x/data/file`
- `go vet ./x/data/file`
