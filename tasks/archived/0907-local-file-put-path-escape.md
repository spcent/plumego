# Card 0907: Local File Put Path Escape

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P1
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/helpers.go
- x/data/file/local.go
- x/data/file/local_test.go
- docs/modules/store/README.md
Depends On:
- 0732

Goal:
Prevent local file uploads from escaping `basePath` through unsafe tenant or generated path components.

Scope:
- Validate upload path components before creating directories or temp files.
- Ensure generated full paths stay inside the storage base directory.
- Add negative tests for traversal-style tenant IDs.

Non-goals:
- Do not move tenant-aware file policy into stable `store/file`.
- Do not redesign file ID generation or metadata ownership.
- Do not add new dependencies.

Files:
- x/data/file/helpers.go
- x/data/file/local.go
- x/data/file/local_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./x/data/file ./store/file
- go test -race -timeout 60s ./x/data/file ./store/file
- go vet ./x/data/file ./store/file

Docs Sync:
- Required for backend safety semantics if behavior becomes explicit.

Done Definition:
- Unsafe tenant/path components cannot escape the local storage root.
- Existing valid uploads continue to work.
- Targeted tests, race tests, and vet pass.

Outcome:
- Added local path component validation for upload tenant IDs.
- Added `safeLocalPath` to validate local storage paths and verify resolved
  filesystem paths stay inside the configured base path.
- Reused the guarded path resolver for local Get/Delete/Exists/Stat/List/Copy.
- Added regression coverage for traversal, absolute, empty, and separator-based
  tenant IDs in `LocalStorage.Put`.
- Documented tenant-aware local backend path isolation requirements.

Validation:
- `go test -timeout 20s ./x/data/file ./store/file`
- `go test -race -timeout 60s ./x/data/file ./store/file`
- `go vet ./x/data/file ./store/file`
