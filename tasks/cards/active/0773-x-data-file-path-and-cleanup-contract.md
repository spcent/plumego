# Card 0773

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/s3.go
- x/data/file/s3_test.go
- x/data/file/local.go
- x/data/file/local_test.go
- docs/modules/x-data/README.md
Depends On:
- 0772-x-data-stable-readiness-fourth-gate

Goal:
Make file provider path validation and metadata-failure cleanup consistent across local and S3 storage.

Scope:
- Apply the documented path safety contract to all S3 public path operations.
- Preserve empty-prefix list behavior while rejecting unsafe non-empty prefixes.
- Surface local cleanup failures when metadata persistence fails after file or thumbnail creation.
- Add focused local and S3 regression tests.

Non-goals:
- Do not add multipart upload.
- Do not change object key generation for successful Put operations.
- Do not change metadata schema.

Files:
- x/data/file/s3.go
- x/data/file/s3_test.go
- x/data/file/local.go
- x/data/file/local_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Update x/data docs with the unified path and cleanup contract.

Done Definition:
- S3 Get/Delete/Exists/Stat/List reject unsafe path-like keys consistently.
- Local metadata-save failure cleanup reports leftover file or thumbnail risk.
- Tests and docs cover the behavior.
