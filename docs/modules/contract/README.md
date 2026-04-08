# contract

## Purpose

`contract` defines structured transport contracts: request context helpers, API errors, and response helpers.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- standardizing error response shape
- writing transport-level response helpers
- carrying request metadata needed by handlers

## Do not use this module for

- protocol gateway families
- business DTO ownership
- route matching
- auth identity contracts
- middleware observability policy or field redaction

## First files to read

- `contract/module.yaml`
- `contract/error.go`
- `contract/response.go`

## Canonical change shape

- preserve one clear error path centered on `NewErrorBuilder` + `WriteError`
- use `Ctx.Response` / `WriteResponse` as the only canonical success path for `Ctx` handlers
- use one explicit bind step per source: `BindJSON(..., BindOptions{...})` for JSON and `BindQuery(...)` for query
- perform validation explicitly via `ValidateStruct(...)` after binding, then write failures through `WriteBindError`
- use `WithRequestID(...)` + `RequestIDFromContext(...)` as the only request-correlation contract; middleware and logging must read from it instead of maintaining package-local request id slots
- keep `RequestIDHeader` as the canonical transport header constant only; request-id attach/read policy belongs to middleware
- keep `TraceContext` for tracing/span state only; do not reuse it as the app-facing request-correlation surface
- use `Ctx.Stream(StreamConfig{...})` as the only high-level streaming/SSE entrypoint; keep `NewSSEWriter(...)` for low-level stdlib-shaped SSE writing only
- keep `WriteJSON` as an explicit lower-level writer for raw payloads outside the `Ctx` success contract
- use stdlib multipart parsing directly in owning handlers such as `x/fileapi`; do not add file-upload or disk-save convenience helpers to `contract.Ctx`
- keep helpers transport-focused
- avoid framework-style abstraction layers
