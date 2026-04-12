# Card 0956: Legacy Response Shim Pruning

Priority: P1
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: multi-module transport helpers
Depends On: —

## Goal

Remove the remaining legacy response helper families that compete with
`contract.WriteResponse`, `contract.WriteError`, and `contract.WriteJSON`, so
Plumego has one canonical HTTP response vocabulary instead of multiple
deprecated shims.

## Problem

- `x/rest/resource.go` still exports deprecated `Response`, `ErrorResponse`,
  `JSONWriter`, and `NewJSONWriter`, even though `x/rest` docs already say
  response and error conventions must align with `contract`.
- `x/rest/BaseResourceController` still carries a `jsonWriter` field and
  helper methods that exist only to route back into `contract`, which leaves a
  second response abstraction in the public API with no repo callers outside
  `resource.go` itself.
- `internal/httpx/response.go` still exports deprecated `Response` and `JSON`
  helpers even though the same package is now only meaningfully used for
  `ClientIP`.
- `x/rest/resource.go` prints response-write failures with `fmt.Printf`, which
  is a side effect inside a library package and further reinforces the stale
  shim path.

## Scope

- Remove deprecated response helper symbols from `x/rest` once all internal
  uses are migrated in the same change.
- Collapse `BaseResourceController` onto direct `contract`-based writes or a
  private helper that does not reintroduce a second public response layer.
- Remove deprecated response helpers from `internal/httpx` while preserving
  `ClientIP`.
- Update `x/rest` docs to point only at `contract` for response and error
  writing.
- Follow the exported-symbol completeness protocol: enumerate all references
  first, migrate every site, and verify the old symbols disappear.

## Non-Goals

- Do not redesign `ResourceSpec`, query parsing, pagination, or CRUD route
  registration.
- Do not change `contract` response shapes.
- Do not change `internal/httpx.ClientIP`.

## Expected Files

- `x/rest/resource.go`
- `x/rest/*_test.go`
- `x/rest/module.yaml`
- `docs/modules/x-rest/README.md`
- `internal/httpx/response.go`

## Validation

```bash
rg -n 'type Response struct|type ErrorResponse struct|type JSONWriter struct|NewJSONWriter|internal/httpx\.JSON|internal/httpx\.Response' x/rest internal . -g '*.go'
go test -timeout 20s ./x/rest/...
go vet ./x/rest/...
```

## Docs Sync

- `docs/modules/x-rest/README.md`
- any `x/rest` package comments that still teach the deprecated helper path

## Done Definition

- `x/rest` no longer exports a deprecated response-helper family that competes
  with `contract`.
- `BaseResourceController` no longer stores or exposes `JSONWriter`.
- `internal/httpx/response.go` no longer exports deprecated response helpers.
- No `fmt.Printf`-based response-write fallback remains in `x/rest`.
- Grep for removed symbols is empty and focused `x/rest` validation passes.

## Outcome

- Removed deprecated `x/rest` response shim types and the `JSONWriter` helper family.
- Collapsed `BaseResourceController` onto direct `contract`-based not-implemented writes.
- Removed `internal/httpx/response.go`; `internal/httpx` now only keeps the remaining non-response helpers.
- Added regression coverage for the default `BaseResourceController` error payload.

## Validation Run

```bash
rg -n 'type Response struct|type ErrorResponse struct|type JSONWriter struct|NewJSONWriter|internal/httpx\.JSON|internal/httpx\.Response' x/rest internal/httpx -g '*.go'
go test -timeout 20s ./x/rest/... ./internal/httpx/...
go vet ./x/rest/... ./internal/httpx/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```
