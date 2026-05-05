# Card 0761

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P2
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/s3.go
- x/data/file/s3_test.go
Depends On:

Goal:
Normalize S3 file backend errors so request, signing, client, and response failures carry file operation context.

Scope:
- Wrap S3 Put/Get/Delete/Stat/List/GetURL request and signing failures in *file.Error where applicable.
- Preserve ErrNotFound/ErrInvalidPath detection via errors.Is.
- Add focused tests for at least one request/signing or non-200 path not already covered.

Non-goals:
- Do not redesign the S3 signer.
- Do not add external dependencies.

Files:
- x/data/file/s3.go
- x/data/file/s3_test.go

Tests:
- go test ./x/data/file

Docs Sync:
- Not required unless public docs mention raw S3 errors.

Done Definition:
- S3 backend error shape is consistent enough for callers to rely on Op/Path and sentinel wrapping.
- S3 file tests pass.

Outcome:
- Wrapped S3 Put/Get/Delete/Stat/List/GetURL request, signing, client, decode,
  and response failures in *storefile.Error where those operations have a file path
  or prefix context.
- Preserved sentinel wrapping for ErrNotFound and ErrInvalidPath.
- Added Stat non-OK response coverage for *storefile.Error shape.
- Validated with:
  - go test -timeout 20s ./x/data/file
  - go test -race -timeout 60s ./x/data/file
  - go vet ./x/data/file
