# 0771 - Core Public Consumer Path Tests

State: done
Priority: P1
Primary module: core

## Goal

Expand `package core_test` coverage for behavior stable consumers rely on without reading private state.

## Scope

- Cover public TLS serve path with configured certificate material.
- Cover public post-shutdown behavior through `Server`, `Prepare`, `Shutdown`, and `ServeHTTP`.
- Cover router policy and named route URL behavior through core public APIs.

## Non-goals

- Do not remove valuable internal tests.
- Do not test private fields from the external package.
- Do not change runtime behavior.

## Files

- `core/core_public_test.go`
- `tasks/cards/active/0771-core-public-consumer-path-tests.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`

## Docs Sync

No docs change expected unless behavior gaps are discovered.

## Done Definition

- External tests cover the listed consumer paths using only public APIs.
- Core targeted tests pass.

## Validation

- `gofmt -w core/core_public_test.go`
- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`
