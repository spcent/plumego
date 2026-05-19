---

State: done
id: 1359
title: x/data file S3 large object policy
status: done
priority: P2
primary_module: x/data/file
---

## Goal

Define and enforce the current S3 provider's large-object policy instead of
leaving single PUT spooling as an unstated production limitation.

## Scope

- Add explicit S3 upload size policy to configuration.
- Reject uploads that exceed the configured single PUT limit before object
  creation when size is known.
- Enforce the same limit while streaming unknown-sized readers to temp files.
- Document that multipart upload is not yet implemented.

## Non-goals

- Implement S3 multipart upload.
- Change local storage upload limits.
- Add non-stdlib dependencies.

## Files

- `x/data/file/config.go`
- `x/data/file/s3.go`
- `x/data/file/s3_test.go`
- `docs/modules/x-data/README.md`

## Tests

- `go test -timeout 20s ./x/data/file`
- `go test -race -timeout 60s ./x/data/file`
- `go vet ./x/data/file`

## Docs Sync

Update `docs/modules/x-data/README.md` stable blockers and S3 provider notes.

## Done Definition

- Oversized S3 uploads return a clear error.
- Unknown-size readers cannot exceed the configured spool limit silently.
- Docs state the current single PUT limit and multipart blocker.

## Validation

- `go test -timeout 20s ./x/data/file`
- `go test -race -timeout 60s ./x/data/file`
- `go vet ./x/data/file`
