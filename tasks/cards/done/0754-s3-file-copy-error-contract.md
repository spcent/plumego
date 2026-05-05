# Card 0754

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P2
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/s3.go
- x/data/file/s3_test.go
Depends On:

Goal:
Make S3Storage.Copy expose missing-source and backend failures through the store/file error contract.

Scope:
- Map 404 copy responses to store/file.ErrNotFound.
- Wrap copy failures in *store/file.Error with operation and path context.
- Add regression coverage for missing copy source.

Non-goals:
- Do not implement AWS-specific multipart copy behavior.
- Do not change S3 URL signing.

Files:
- x/data/file/s3.go
- x/data/file/s3_test.go

Tests:
- go test -timeout 20s ./x/data/file

Docs Sync:
- None.

Done Definition:
- errors.Is(err, storefile.ErrNotFound) works for missing S3 copy source.
- Other S3 copy failures include file operation context.
- Targeted tests pass.

Outcome:
S3Storage.Copy now wraps request/signing/client/status failures in *storefile.Error. A 404 copy response maps to ErrNotFound with source path context. Added missing-source regression coverage.

Validation:
- go test -timeout 20s ./x/data/file
