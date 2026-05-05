# Card 0745

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P2
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/s3.go
- x/data/file/s3_test.go
Depends On:
- tasks/cards/active/0740-s3-file-tenant-key-validation.md

Goal:
Build S3 object URLs without escaping path separators or normalizing unsafe keys.

Scope:
- Escape S3 object key segments instead of the whole key.
- Preserve slash separators in path-style and virtual-hosted URLs.
- Keep invalid-path validation aligned with card 0740.
- Add tests for nested object keys and unsafe key rejection.

Non-goals:
- Do not implement full AWS signing.
- Do not change bucket/endpoint configuration behavior.

Files:
- x/data/file/s3.go
- x/data/file/s3_test.go

Tests:
- go test -timeout 20s ./x/data/file

Docs Sync:
- None unless docs mention URL shape.

Done Definition:
- Nested S3 keys render as path segments, not `%2F`.
- Unsafe keys are rejected before URL construction.
- Targeted tests pass.

Outcome:

