# Card 0739

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P1
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/types.go
- x/data/file/metadata.go
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/*_test.go
Depends On:

Goal:
Prevent file hash deduplication from returning metadata across tenant boundaries.

Scope:
- Make dedupe lookup tenant-aware.
- Update Local and S3 Put to query dedupe records for the current tenant only.
- Preserve existing metadata manager behavior for non-tenant callers where safe.
- Add regression tests for same hash across different tenants.

Non-goals:
- Do not introduce cross-tenant shared blob ownership semantics.
- Do not change stable store/file interfaces.

Files:
- x/data/file/types.go
- x/data/file/metadata.go
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/local_test.go
- x/data/file/metadata_test.go
- x/data/file/s3_test.go

Tests:
- go test -timeout 20s ./x/data/file

Docs Sync:
- Update docs only if dedupe semantics are documented.

Done Definition:
- A tenant cannot receive another tenant's metadata through hash dedupe.
- Existing same-tenant dedupe behavior remains covered.
- Targeted tests pass.

Outcome:

