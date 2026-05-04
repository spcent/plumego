# Card 0736

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/helpers.go
- x/data/file/local.go
- x/data/file/s3.go
- x/data/file/local_test.go
- x/data/file/s3_test.go
- docs/modules/x-data/README.md
Depends On:
- 0735-x-data-kvengine-wal-config-stability

Goal:
Close remaining file/S3 readiness gaps and synchronize x/data documentation with implemented APIs.

Scope:
- Surface crypto-random ID generation failures instead of ignoring entropy errors.
- Preserve S3 object-key path separators while escaping unsafe path segments.
- Escape S3 copy-source keys correctly.
- Fix non-compiling sharding quick-start documentation.

Non-goals:
- Do not redesign S3 multipart streaming.
- Do not add new providers.
- Do not move file metadata ownership.

Files:
- x/data/file/helpers.go
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
- Update `docs/modules/x-data/README.md` quick start and support notes to match code.

Done Definition:
- File IDs fail closed if secure random generation fails.
- S3 URL/copy path handling preserves object hierarchy and escapes unsafe segments.
- x/data docs snippets match current public API.

Outcome:
- Made file ID generation return secure-random read errors and surfaced them from local/S3 `Put`.
- Changed S3 object-key escaping to preserve `/` hierarchy while escaping unsafe path segments and traversal markers.
- Escaped `x-amz-copy-source` by path segment.
- Updated x/data docs with file ID and S3 path-safety behavior.

Validation:
- GOCACHE=/private/tmp/plumego-go-build go test -timeout 20s ./x/data/file
- GOCACHE=/private/tmp/plumego-go-build go test -race -timeout 60s ./x/data/file
- GOCACHE=/private/tmp/plumego-go-build go vet ./x/data/file
