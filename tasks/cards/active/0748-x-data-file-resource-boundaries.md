# Card 0748

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/config.go
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/s3_test.go
- docs/modules/x-data/README.md
Depends On:
- 0747-x-data-idempotency-dialect-duplicate-policy

Goal:
Make file storage resource boundaries explicit for S3 spooling, error bodies, and local durable rename.

Scope:
- Add S3Config.TempDir for upload spooling.
- Add bounded S3 error-body reads.
- Fsync the local parent directory after rename where supported.
- Add tests for TempDir spooling and bounded error-body output.

Non-goals:
- Do not add multipart upload support.
- Do not add cloud SDK dependencies.
- Do not change MetadataManager contracts.

Files:
- x/data/file/config.go
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/s3_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Update x/data docs for S3 temp dir, bounded error reads, and local rename durability.

Done Definition:
- S3 upload spooling location can be configured.
- S3 error messages do not read unbounded response bodies.
- Local Put syncs the containing directory after rename where possible.

Outcome:
