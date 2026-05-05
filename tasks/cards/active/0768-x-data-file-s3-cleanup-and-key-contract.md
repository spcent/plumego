# Card 0768

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/s3.go
- x/data/file/s3_test.go
- docs/modules/x-data/README.md
Depends On:
- 0767-x-data-file-local-list-cancellation

Goal:
Make S3 object cleanup and key validation semantics explicit and consistent.

Scope:
- Surface cleanup failure when metadata persistence fails after object upload.
- Reject unsafe path-like keys consistently for URL and copy operations, or document object-key-only semantics with tests.
- Add focused tests for cleanup error reporting and key validation.

Non-goals:
- Do not add multipart upload.
- Do not change presigned URL signing.
- Do not change the metadata schema.

Files:
- x/data/file/s3.go
- x/data/file/s3_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Update x/data docs with S3 cleanup and key contract.

Done Definition:
- Orphan-object cleanup failures are visible to callers.
- S3 key handling is consistent with the documented file path contract.
- Tests and docs cover the behavior.

