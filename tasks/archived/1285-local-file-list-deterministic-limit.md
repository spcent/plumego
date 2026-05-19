# Card 1285

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/file
Owned Files: x/data/file/local.go, x/data/file/local_test.go
Depends On:

Goal:

Make LocalStorage.List apply limit after collecting and sorting so callers receive a deterministic lexicographic prefix of the result set.

Scope:

- Collect all matching local file metadata before applying limit.
- Sort by Path before truncating.
- Support empty prefix as the storage root.
- Keep invalid path checks and tenant path isolation intact.
- Cover deterministic limit behavior in tests.

Non-goals:

- Adding pagination cursors.
- Changing metadata schema.
- Optimizing large-directory scans beyond deterministic behavior.

Files:

- x/data/file/local.go
- x/data/file/local_test.go

Tests:

- go test -race -timeout 60s ./x/data/file/...
- go test -timeout 20s ./x/data/file/...
- go vet ./x/data/file/...

Docs Sync:

- Not required; this fixes implementation to match the existing sorted-list behavior.

Done Definition:

- Limit is applied only after sorting by Path.
- Tests cover stable sorted truncation.
- Module tests and vet pass.

Outcome:

- LocalStorage.List now supports empty prefix by walking the storage root.
- List now sorts by Path before applying a positive limit.
- Added coverage for empty-prefix listing and sorted limit truncation.
- Validation passed:
  - go test -race -timeout 60s ./x/data/file/...
  - go test -timeout 20s ./x/data/file/...
  - go vet ./x/data/file/...
