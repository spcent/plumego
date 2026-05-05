# Card 0754

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/s3.go
- x/data/file/s3_signer.go
- x/data/file/s3_test.go
- x/data/file/s3_signer_test.go
- docs/modules/x-data/README.md
Depends On:
- 0753-x-data-kvengine-wal-durability-contract

Goal:
Harden S3-compatible signing and status handling so generated URLs and object existence checks fail explicitly.

Scope:
- Ensure presigned canonical requests include the real Host header.
- Use AWS-compatible canonical query escaping for signing.
- Return errors for non-OK/non-404 Exists responses.
- Normalize S3 Stat/List/Copy error shape through store/file errors where useful.
- Add tests for host signing, space encoding, and server-error Exists behavior.

Non-goals:
- Do not add a cloud SDK dependency.
- Do not add multipart upload.
- Do not change object key layout.

Files:
- x/data/file/s3.go
- x/data/file/s3_signer.go
- x/data/file/s3_test.go
- x/data/file/s3_signer_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Document S3 status handling and presigned URL compatibility boundary.

Done Definition:
- Presigned URLs sign the actual host.
- Canonical query encoding is AWS-compatible for spaces.
- S3 server/auth errors are not reported as missing objects.

Outcome:
- Presigned requests now set and sign the URL host.
- Canonical query signing now uses AWS-compatible space encoding and sorted values.
- S3 Exists, Stat, List, and Copy return explicit storage errors for server/auth failures.
- Added regression coverage for presign host handling, AWS query escaping, and Exists 500 behavior.

Validation:
- `go test -timeout 20s ./x/data/file`
- `go test -race -timeout 60s ./x/data/file`
- `go vet ./x/data/file`
