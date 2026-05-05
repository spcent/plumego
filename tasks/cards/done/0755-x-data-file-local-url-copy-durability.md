# Card 0755

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/local.go
- x/data/file/local_test.go
- docs/modules/x-data/README.md
Depends On:
- 0754-x-data-file-s3-presign-and-errors

Goal:
Make local storage URL generation, copy, and thumbnail writes safe and durable enough for production use.

Scope:
- Validate and URL-escape local storage paths in GetURL.
- Make Copy use temp-file write, sync, close, rename, and directory sync.
- Make thumbnail writes use the same atomic local write pattern.
- Add focused tests for escaped URLs and copy durability cleanup.

Non-goals:
- Do not add image processing dependencies.
- Do not change the store/file interface.
- Do not redesign local path layout.

Files:
- x/data/file/local.go
- x/data/file/local_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/file
- go test -race -timeout 60s ./x/data/file
- go vet ./x/data/file

Docs Sync:
- Document local URL escaping and atomic copy behavior.

Done Definition:
- GetURL rejects unsafe paths and returns escaped static URLs.
- Copy and thumbnail writes do not leave acknowledged partial files.
- Tests cover the new behavior.

Outcome:
- Local GetURL now rejects unsafe paths and escapes path segments.
- Local Copy now writes through temp file, sync, close, rename, and directory sync.
- Thumbnail generation now uses the same atomic local write helper.
- Added focused URL escaping and unsafe-copy regression tests.

Validation:
- `go test -timeout 20s ./x/data/file`
- `go test -race -timeout 60s ./x/data/file`
- `go vet ./x/data/file`
