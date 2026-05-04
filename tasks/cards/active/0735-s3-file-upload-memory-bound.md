# Card 0735: S3 File Upload Memory Bound

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P2
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/config.go
- x/data/file/s3.go
- x/data/file/s3_test.go
- docs/modules/store/README.md
Depends On:
- 0734

Goal:
Prevent S3 uploads from unbounded in-memory buffering.

Scope:
- Add or use explicit max upload size configuration for S3 uploads.
- Reject oversized uploads before buffering when `PutOptions.Size` is known.
- Bound reader consumption when size is unknown.
- Add tests for known and unknown oversized upload rejection.

Non-goals:
- Do not add multipart streaming or AWS SDK dependencies.
- Do not change stable `store/file.Storage`.
- Do not move provider config into stable `store/file`.

Files:
- x/data/file/config.go
- x/data/file/s3.go
- x/data/file/s3_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./x/data/file ./store/file
- go test -race -timeout 60s ./x/data/file ./store/file
- go vet ./x/data/file ./store/file

Docs Sync:
- Required for provider memory envelope.

Done Definition:
- S3 Put has a deterministic memory bound.
- Oversized payloads return `store/file.ErrInvalidSize`.
- Targeted tests, race tests, and vet pass.

Outcome:
