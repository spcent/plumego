# Card 1245

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P2
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/local.go
- x/data/file/local_test.go
Depends On:

Goal:
Make Local Exists and Stat return the same file.Error shape as other Local operations.

Scope:
- Wrap invalid path, missing file, and stat errors with Op/Path context.
- Preserve errors.Is behavior for ErrInvalidPath and ErrNotFound.
- Add focused tests for Exists and Stat error shape.

Non-goals:
- Do not change Local Put/Get/Delete/Copy behavior beyond shared helpers if needed.

Files:
- x/data/file/local.go
- x/data/file/local_test.go

Tests:
- go test ./x/data/file

Docs Sync:
- Not required unless public docs mention raw errors.

Done Definition:
- Local Exists/Stat match the backend's established file.Error contract.
- Local file tests pass.

Outcome:
- Wrapped LocalStorage.Exists invalid-path and stat failures in *storefile.Error.
- Wrapped LocalStorage.Stat invalid-path, not-found, and stat failures in
  *storefile.Error.
- Added focused error-shape tests for Exists and Stat.
- Validated with:
  - go test -timeout 20s ./x/data/file
  - go test -race -timeout 60s ./x/data/file
  - go vet ./x/data/file
