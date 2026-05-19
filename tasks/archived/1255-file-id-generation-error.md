# Card 1255

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P3
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/helpers.go
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/local_test.go
- x/data/file/s3_test.go
Depends On:

Goal:
Make file ID generation fail explicitly when crypto/rand cannot provide entropy.

Scope:
- Change the internal generateID helper to return an error.
- Make Local and S3 Put fail with file operation context when ID generation fails.
- Add tests through an injectable or internal-only seam if needed.

Non-goals:
- Do not change exported file storage APIs.
- Do not add external ID generation dependencies.

Files:
- x/data/file/helpers.go
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/local_test.go
- x/data/file/s3_test.go

Tests:
- go test ./x/data/file

Docs Sync:
- Not required unless public docs mention ID generation guarantees.

Done Definition:
- crypto/rand errors are no longer ignored.
- File backend tests pass.

Outcome:
- Changed generateID to return crypto/rand errors instead of ignoring them.
- Updated LocalStorage.Put and S3Storage.Put to return *storefile.Error when ID
  generation fails before a storage path is created.
- Added Local and S3 Put regression tests for entropy failure.
- Validated with:
  - go test -timeout 20s ./x/data/file
  - go test -race -timeout 60s ./x/data/file
  - go vet ./x/data/file
