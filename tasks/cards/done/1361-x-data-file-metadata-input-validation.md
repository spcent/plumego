---
id: 1361
title: x/data file metadata input validation
status: done
priority: P2
primary_module: x/data/file
---

## Goal

Make direct `MetadataManager` use enforce the same tenant and path validation
that storage-backed calls already rely on.

## Scope

- Validate nil file, tenant id, path, id, and hash inputs where applicable.
- Clean tenant IDs consistently in `Save`, `GetByHash`, and `List`.
- Add tests for invalid tenant/path/list behavior.

## Non-goals

- Change the metadata schema.
- Add admin list behavior beyond existing `ListAll`.
- Change storage provider dedupe behavior.

## Files

- `x/data/file/metadata.go`
- `x/data/file/metadata_test.go`

## Tests

- `go test -timeout 20s ./x/data/file`
- `go test -race -timeout 60s ./x/data/file`
- `go vet ./x/data/file`

## Docs Sync

No docs change expected unless public error behavior needs mention.

## Done Definition

- Direct metadata calls fail closed on invalid tenant/path inputs.
- `List` cannot bypass tenant validation with raw tenant strings.
- Tests cover invalid direct manager usage.

## Validation

- `go test -timeout 20s ./x/data/file`
- `go test -race -timeout 60s ./x/data/file`
- `go vet ./x/data/file`
