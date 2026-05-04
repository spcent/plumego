# Card 0740

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/local_test.go
- x/data/file/s3_test.go
- docs/modules/x-data/README.md
Depends On:
- 0739-x-data-kvengine-config-and-observability

Goal:
Make file storage memory and durability behavior explicit enough for stable use.

Scope:
- Handle local temporary file close errors before rename.
- Add local fsync before rename for durable writes where supported.
- Avoid unbounded S3 upload buffering by spooling upload content to a temporary file while hashing.
- Add tests for close/write error paths and S3 request body behavior.

Non-goals:
- Do not implement multipart uploads.
- Do not change the metadata store interface.
- Do not add cloud SDK dependencies.

Files:
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/local_test.go
- x/data/file/s3_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Update x/data docs for local durability and S3 spooling behavior.

Done Definition:
- Local Put does not ignore close errors before rename.
- S3 Put does not keep the whole object in memory for upload.
- Docs describe implemented large-object behavior only.

Outcome:
