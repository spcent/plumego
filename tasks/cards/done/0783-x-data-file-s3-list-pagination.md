---
id: 0783
title: x/data file S3 list pagination
status: done
priority: P1
primary_module: x/data/file
---

## Goal

Make `S3Storage.List` consume S3 `ListObjectsV2` pagination so callers do not
silently receive partial results from large buckets.

## Scope

- Parse `IsTruncated` and `NextContinuationToken`.
- Continue listing until the requested limit is reached or S3 reports no more
  pages.
- Preserve existing `List(ctx, prefix, limit)` API.
- Add tests for multi-page listing and limit handling.

## Non-goals

- Add a new cursor-returning public API.
- Implement delimiter or common-prefix semantics.
- Change local storage listing behavior.

## Files

- `x/data/file/s3.go`
- `x/data/file/s3_test.go`

## Tests

- `go test -timeout 20s ./x/data/file`
- `go test -race -timeout 60s ./x/data/file`
- `go vet ./x/data/file`

## Docs Sync

Update docs only if the List contract wording changes.

## Done Definition

- Truncated S3 list responses are followed using continuation tokens.
- `limit` bounds total returned files across pages.
- Existing non-paginated tests continue to pass.

## Validation

- `go test -timeout 20s ./x/data/file`
- `go test -race -timeout 60s ./x/data/file`
- `go vet ./x/data/file`
