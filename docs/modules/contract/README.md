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
- `contract/errors.go`
- `contract/response.go`
- `contract/bind_helpers.go`
- `contract/context_bind.go`
- `contract/context_core.go`
- `contract/request_id.go`
- `contract/trace.go`

## Public entrypoints

- `HeaderContentType`
- `ContentTypeJSON`
- `Response`
- `NewErrorBuilder`
- `ErrorBuilder`
- `ErrorCategory`
- `ErrorSeverity`
- `ErrorType`
- `ErrorTypeMeta`
- `ErrorType.Meta`
- `CategoryForStatus`
- `HTTPStatusFromCategory`
- `APIError`
- `ErrorResponse`
- `Code*` transport error constants
- `WriteError`
- `WriteResponse`
- `WriteJSON`
- `WriteBindError`
- `BindErrorToAPIError`
- `BindOptions`
- `FieldError`
- `FieldErrorsFrom`
- `ValidationErrors`
- `ValidateStruct`
- `NewCtx`
- `NewCtxWithConfig`
- `Ctx`
- `RequestConfig`
- `Ctx.Param`
- `Ctx.MustParam`
- `Ctx.RequestHeaders`
- `Ctx.RequestID`
- `WithRequestContext`
- `RequestContextFromContext`
- `RoutePatternFromContext`
- `RouteNameFromContext`
- `WithRequestID`
- `RequestIDFromContext`
- `RequestIDHeader`
- `TraceID`
- `SpanID`
- `TraceFlags`
- `TraceFlagsSampled`
- `TraceIDLength`
- `SpanIDLength`
- `WithTraceContext`
- `TraceContextFromContext`
- `WithSpanIDString`
- `TraceContext`
- `TraceContext.IsSampled`
- `ParseTraceID`
- `ParseSpanID`
- `IsValidTraceID`
- `IsValidSpanID`
- `Err*` transport sentinel errors

## Canonical change shape

- preserve one clear error path centered on `NewErrorBuilder` + `WriteError`
- pass canonical `Code*` constants or uppercase stable strings to `ErrorBuilder.Code`; the builder preserves explicit caller input
- rely on `WriteError`/`NewErrorBuilder` to fill missing codes with canonical machine-safe `Code*` constants, never title-cased HTTP reason phrases
- use `WriteResponse` as the canonical success response path
- use one explicit bind step per source: `BindJSON(..., BindOptions{...})` for JSON and `BindQuery(...)` for query
- perform validation explicitly via `ValidateStruct(...)` after binding, then write failures through `WriteBindError`
- treat unknown or malformed validation rules as `ErrValidationConfig` programmer errors; `WriteBindError` maps those to server errors, not client validation failures
- use `WithRequestID(...)` + `RequestIDFromContext(...)` as the only request-correlation contract; middleware and logging must read from it instead of maintaining package-local request id slots
- keep request-id generation policy in `middleware/requestid` or middleware-owned observability helpers, not in `contract`
- keep `RequestIDHeader` as the canonical transport header constant only; request-id attach/read policy belongs to middleware
- keep `TraceContext` for tracing/span state only; it is a defensive context carrier, not a tracing runtime or propagation implementation
- keep transport helpers deterministic and side-effect-free; do not add package-global warning or diagnostics hooks
- keep `Ctx` as a legacy compatibility carrier for `http.ResponseWriter`, `*http.Request`, route params, and narrow binding helpers only; do not add response-writing helpers, string-key request bags, abort state, hidden deadlines, or request-local service-locator helpers
- treat `Ctx.BindJSON` as a cache-oriented compatibility helper: it reads the body into memory once and restores `R.Body` only when body cache is enabled
- treat `BindOptions.MaxBodySize` as an endpoint-specific post-read cap; read-time protection belongs to `RequestConfig.MaxBodySize`
- keep protocol-specific streaming/SSE helpers out of stable `contract`; owning modules should implement those on explicit `net/http` handlers
- keep `WriteJSON` as an explicit lower-level writer for raw payloads outside the `Ctx` success contract
- use stdlib multipart parsing directly in owning handlers such as `x/fileapi`; do not add file-upload or disk-save convenience helpers to `contract.Ctx`
- keep generic internal error wrapping, panic recovery helpers, and HTTP response parsing local to the owning module; do not add repo-wide error utility helpers to `contract`
- keep helpers transport-focused
- avoid framework-style abstraction layers

## Frozen behavior matrix

These behaviors are part of the current stable-root freeze baseline:

| Surface | Behavior |
| --- | --- |
| `WriteResponse` | writes the canonical success envelope and injects `request_id` from context when present |
| `WriteResponse` / `WriteJSON` | statuses without bodies write headers only and do not set JSON content type |
| `WriteResponse` / `WriteJSON` | invalid success statuses normalize to `500` before writing |
| `WriteResponse` | `nil` data and `nil` meta produce the empty JSON object `{}` for body-eligible statuses |
| `WriteError` | writes the canonical error envelope and uses top-level `request_id` only |
| `WriteError` | incomplete or invalid `APIError` values are normalized deterministically |
| `NewErrorBuilder().Type(...)` | applies canonical status, code, and category for the selected type |
| `ErrorType.Meta()` | returns the nameable `ErrorTypeMeta` value for the selected type |
| `NewErrorBuilder().TypeOnly(...)` | records type while preserving explicit status, code, and category |
| `Details(...)` / `Detail(...)` | clone detail maps and omit empty detail keys |
| `Ctx.BindJSON` | reads and optionally caches request body bytes before decoding |
| `BindOptions.MaxBodySize` | enforces a stricter post-read cap after `RequestConfig.MaxBodySize` read-time protection |
| `TraceContext` | stores trace/span metadata defensively but does not parse or inject propagation headers |
| Nil response writer | `WriteJSON`, `WriteResponse`, and `WriteError` return `ErrResponseWriterNil` |

Focused regression coverage lives in `contract/freeze_test.go`,
`contract/errors_test.go`, and `contract/active_cards_regression_test.go`.

## Stable Semantics Notes

`TraceContext` is stable as a transport metadata carrier only. It defensively
copies baggage and parent span ids, accepts caller-provided values without
header propagation, and leaves sampling, extraction, injection, and collector
policy to `x/observability`.

`Ctx.BindJSON` is stable as a legacy compatibility helper. It reads the body
into memory once, optionally restores `R.Body` for later readers, and applies
`BindOptions.MaxBodySize` only after read-time protection from
`RequestConfig.MaxBodySize`.

`WriteResponse` and `WriteJSON` preserve HTTP no-body semantics. `204`, `304`,
and informational statuses write headers only; body-eligible success responses
use the canonical envelope, which encodes as `{}` when every envelope field is
empty.
