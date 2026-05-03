# Card 0726

Milestone:
Recipe: specs/change-recipes/tenant-policy-change.yaml
Priority: P0
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/types.go
- x/data/file/helpers.go
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/metadata.go
Depends On:

Goal:
Prevent tenant isolation bypasses in x/data/file storage and deduplication.

Scope:
- Validate tenant IDs before deriving local filesystem paths or S3 object keys.
- Ensure deduplication lookup is tenant-scoped instead of global hash-only lookup.
- Preserve existing storage interfaces where possible and keep HTTP behavior out of x/data/file.
- Add focused tests for invalid tenant path input and cross-tenant same-hash behavior.

Non-goals:
- Do not redesign the file metadata schema beyond query changes needed for tenant-scoped lookup.
- Do not add a new storage backend.
- Do not move HTTP upload/download behavior into this module.

Files:
- x/data/file/types.go
- x/data/file/helpers.go
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/metadata.go
- x/data/file/*_test.go

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Update docs/modules/x-data/README.md only if the public metadata contract changes.

Done Definition:
- Invalid tenant IDs cannot shape filesystem or S3 paths.
- Same-content uploads in different tenants do not return another tenant's metadata.
- x/data/file tests and vet pass.

Outcome:
