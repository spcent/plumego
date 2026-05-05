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
- `TraceContext.HasTraceID`
- `TraceContext.HasSpanID`
- `TraceContext.Valid`
- `ParseTraceID`
- `ParseSpanID`
- `IsValidTraceID`
- `IsValidSpanID`
- `Err*` transport sentinel errors

## Public Surface Freeze

The v1 support surface is intentionally split into three groups:

| Group | Entrypoints | Stable meaning |
| --- | --- | --- |
| Core stable transport contracts | `Response`, `WriteResponse`, `WriteJSON`, `WriteError`, `ErrorResponse`, `APIError`, `NewErrorBuilder`, `ErrorBuilder`, `ErrorCategory`, `ErrorSeverity`, `ErrorType`, `ErrorTypeMeta`, `ErrorType.Meta`, `CategoryForStatus`, `HTTPStatusFromCategory`, `Code*`, `WriteBindError`, `BindErrorToAPIError` | Canonical success and error response shape. Behavior changes require explicit stable-root review. |
| Request metadata carriers | `WithRequestID`, `RequestIDFromContext`, `WithRequestContext`, `RequestContextFromContext`, `RoutePatternFromContext`, `RouteNameFromContext`, `WithTraceContext`, `TraceContextFromContext`, `WithSpanIDString`, `TraceContext`, trace validity helpers | Defensive transport metadata carriers only. They do not own routing, auth, tenant, tracing, or observability policy. |
| Compatibility helpers | `BindOptions`, `FieldError`, `FieldErrorsFrom`, `ValidationErrors`, `ValidateStruct`, `NewCtx`, `NewCtxWithConfig`, `Ctx`, `RequestConfig`, `Ctx.Param`, `Ctx.MustParam`, `Ctx.RequestHeaders`, `Ctx.RequestID` | Retained for compatibility and narrow convenience. They are supported, but not preferred expansion points for new handler styles. |
| Guardrail constants and scalar types | `HeaderContentType`, `ContentTypeJSON`, `RequestIDHeader`, `TraceID`, `SpanID`, `TraceFlags`, `TraceFlagsSampled`, `TraceIDLength`, `SpanIDLength`, `Err*` sentinels | Shared transport names and sentinel values. Additions must stay transport-owned and stdlib-compatible. |

Future work that narrows `Ctx`, binding helpers, `ValidateStruct`, exported
`APIError` fields, or context carrier fields is breaking work. It must use a
dedicated symbol-change card with full caller enumeration before implementation.

## Error Taxonomy

`ErrorType.Meta()` is the canonical taxonomy lookup. It owns the default
`Status`, `Category`, and `Code` for every public error type. Builder and writer
normalization must converge back to this table for status and category.

| Category | Meaning | Representative statuses |
| --- | --- | --- |
| `validation_error` | Request validation and bind input failures | `400` |
| `auth_error` | Authentication and authorization failures | `401`, `403` |
| `client_error` | Generic client input and resource-state failures | `400`, `404`, `405`, `409`, `410`, `422` |
| `rate_limit_error` | Throttling failures | `429` |
| `timeout_error` | Request timeout and upstream gateway timeout failures | `408`, `504` |
| `server_error` | Infrastructure, programming, upstream, and unavailable service failures | `500`, `501`, `502`, `503` |

Extension-owned error codes may refine a canonical `ErrorType`, but they must
stay in the same status and category family. Extension codes should be stable
uppercase machine names owned by the extension module. Do not add business
categories, extension-specific constants, or feature policy to `contract`.

## Canonical change shape

- preserve one clear error path centered on `NewErrorBuilder` + `WriteError`
- build new `APIError` values with `NewErrorBuilder`; direct `APIError` literals are a compatibility path and are normalized by `WriteError`
- keep non-test callers outside `contract` on `NewErrorBuilder`; `go test ./contract` rejects external `contract.APIError{}` literals
- pass canonical `Code*` constants or uppercase stable strings to `ErrorBuilder.Code`; the builder preserves explicit caller input
- do not pair one `contract.Type*` with an incompatible `contract.Code*`; `go test ./contract` rejects contract-owned code overrides across type families
- keep extension-owned custom codes in the same status and category family as the selected `ErrorType`
- rely on `WriteError`/`NewErrorBuilder` to fill missing codes with canonical machine-safe `Code*` constants, never title-cased HTTP reason phrases
- use `TypeTimeout` for request timeout responses and `TypeGatewayTimeout` for upstream gateway timeout responses
- use `WriteResponse` as the canonical success response path
- use one explicit bind step per source: `BindJSON(..., BindOptions{...})` for JSON and `BindQuery(...)` for query
- perform validation explicitly via `ValidateStruct(...)` after binding, then write failures through `WriteBindError`
- treat `ValidateStruct(...)` as a retained compatibility validator for simple struct tags, not a general validation framework; new complex policy validation belongs in the owning module
- treat bind destination, nil context/request, and bind option errors as programmer/configuration failures; `WriteBindError` maps them to internal errors
- treat unknown or malformed validation rules as `ErrValidationConfig` programmer errors; `WriteBindError` maps those to server errors, not client validation failures
- use `WithRequestID(...)` + `RequestIDFromContext(...)` as the only request-correlation contract; middleware and logging must read from it instead of maintaining package-local request id slots
- keep request-id generation policy in `middleware/requestid` or middleware-owned observability helpers, not in `contract`
- keep `RequestIDHeader` as the canonical transport header constant only; request-id attach/read policy belongs to middleware
- keep `TraceContext` for tracing/span state only; it is a defensive context carrier, not a tracing runtime or propagation implementation
- keep transport helpers deterministic and side-effect-free; do not add package-global warning or diagnostics hooks
- keep `Ctx` as a legacy compatibility carrier for `http.ResponseWriter`, `*http.Request`, route params, and narrow binding helpers only; do not add response-writing helpers, string-key request bags, abort state, hidden deadlines, or request-local service-locator helpers
- treat `Ctx.BindJSON` as a cache-oriented compatibility helper: it reads the body into memory once and restores `R.Body` only when body cache is enabled
- keep `BindQuery` limited to primitive scalar values, primitive slices, and scalar types implementing `encoding.TextUnmarshaler`
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
| `WriteResponse` | preserves any valid caller-selected HTTP status, including non-2xx statuses used by health/readiness style JSON bodies |
| `WriteResponse` / `WriteJSON` | statuses without bodies write headers only and do not set JSON content type |
| `WriteResponse` / `WriteJSON` | invalid success statuses normalize to `500` before writing |
| `WriteResponse` | `nil` data and `nil` meta produce the empty JSON object `{}` for body-eligible statuses |
| `WriteError` | writes the canonical error envelope and uses top-level `request_id` only |
| `WriteError` | incomplete or invalid `APIError` literals are normalized deterministically as compatibility behavior |
| `NewErrorBuilder().Type(...)` | applies canonical status, code, and category for the selected type; custom codes may override the default code, but status and category remain canonical |
| Invalid `APIError.Type` / `APIError.Severity` | unrecognized values are omitted during normalization rather than being emitted on the wire |
| `ErrorType.Meta()` | returns the nameable `ErrorTypeMeta` value for the selected type; unknown types fail closed to internal server error metadata |
| `Details(...)` / `Detail(...)` | clone detail maps and omit empty detail keys |
| `Ctx.BindJSON` | reads and optionally caches request body bytes before decoding |
| `BindOptions.MaxBodySize` | enforces a stricter post-read cap after `RequestConfig.MaxBodySize` read-time protection |
| `TraceContext` | stores trace/span metadata defensively and exposes validity helpers, but does not parse or inject propagation headers |
| Nil response writer | `WriteJSON`, `WriteResponse`, and `WriteError` return `ErrResponseWriterNil` |

Focused regression coverage lives in `contract/freeze_test.go`,
`contract/errors_test.go`, and `contract/active_cards_regression_test.go`.

## Stable Readiness Gates

Before treating `contract` changes as release-ready, run:

```bash
go test -race -timeout 60s ./contract/...
go test -timeout 20s ./contract/...
go vet ./contract/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests
go run ./internal/checks/agent-workflow
```

For code-bearing changes that affect public behavior or cross-module callers,
also run `go test -timeout 20s ./...` and `go vet ./...` before handoff.

## Stable Semantics Notes

`Ctx`, `BindJSON`, `BindQuery`, and `ValidateStruct` remain in the stable
surface for compatibility. They are not the preferred expansion point for new
endpoint styles. New stable transport code should keep handlers on
`func(http.ResponseWriter, *http.Request)`, decode with explicit stdlib
operations where practical, and keep module-specific validation policy in the
owning package. Removing or narrowing these helpers is future breaking work and
must go through a dedicated symbol-change card with full caller enumeration.
`ValidateStruct` supports only the documented compatibility rules:
`required`, `email`, `min=<n>`, and `max=<n>`, plus recursive traversal of
exported nested structs, slices, and arrays. Nil inputs, nil non-struct
pointers, and non-struct values are no-ops. `required` uses Go zero-value
semantics after pointer/interface indirection: empty strings, false booleans,
zero numbers, nil pointers, and empty slices/maps/arrays fail. Validation stops
at depth 10 and reports an out-of-range field error for deeper structures. New
rule names, localization, cross-field validation, conditional validation, and
business policy checks belong in the owning module or a dedicated extension.

The retained validation rule/type matrix is:

| Rule | Supported targets | Unsupported target behavior |
| --- | --- | --- |
| `required` | all kinds after pointer/interface indirection | zero values fail |
| `email` | strings | non-strings fail with `INVALID_FORMAT`; empty strings pass unless also `required` |
| `min=<n>` | strings, signed/unsigned integers, floats | other kinds are no-ops |
| `max=<n>` | strings, signed/unsigned integers, floats | other kinds are no-ops |

Malformed `min`/`max` arguments and unknown rule names are
`ErrValidationConfig` programmer errors.

`TraceContext` is stable as a transport metadata carrier only. It defensively
copies baggage and parent span ids, accepts caller-provided values without
header propagation, and leaves sampling, extraction, injection, and collector
policy to `x/observability`. `TraceContext.Baggage` is copied carrier data; key
validation, size limits, extraction, injection, and propagation policy are not
owned by `contract`. Adding policy fields or propagation behavior to
`TraceContext` is out of scope for the stable root. Callers that retrieve a
trace context must treat it as carrier data until `TraceContext.Valid()` returns
true; invalid identifiers are preserved for inspection, not upgraded into a
trusted tracing context.

`Ctx.BindJSON` is stable as a legacy compatibility helper. It reads the body
into memory once, optionally restores `R.Body` for later readers, and applies
`BindOptions.MaxBodySize` only after read-time protection from
`RequestConfig.MaxBodySize`.

`BindQuery` supports fields with explicit `query` tags for strings, signed and
unsigned integers, floats, booleans, pointers to supported scalar values,
slices of supported scalar values, and scalar values implementing
`encoding.TextUnmarshaler` such as `time.Time`. Repeated scalar parameters use
the first value returned by `url.Values.Get`; slices preserve all repeated
values, including explicit empty values. Missing values leave fields unchanged.
Embedded structs are not recursively flattened, and maps, nested pointers, and
arbitrary nested objects are unsupported bind destinations.

`WriteResponse` and `WriteJSON` preserve HTTP no-body semantics. `204`, `304`,
and informational statuses write headers only; body-eligible success responses
use the canonical envelope, which encodes as `{}` when every envelope field is
empty. This empty object is the stable success-envelope representation for
body-eligible responses with no data, meta, or request id; callers that need no
body should choose a no-body status such as `204`. `WriteResponse` and
`WriteJSON` preserve any valid caller-selected HTTP status because some
transport health/readiness responses intentionally return structured JSON with
non-2xx status codes. Error payloads still belong on the canonical `WriteError`
path, not on `WriteResponse`.

`RequestContext` is stable as router-owned metadata only: params, route pattern,
and route name. Tenant identity, auth identity, session state, feature flags,
service handles, and other request-local runtime objects must use their owning
module's explicit context helpers, not `RequestContext`. Adding new fields to
`RequestContext` is a stable carrier expansion and must go through a dedicated
boundary review.

`WriteBindError` separates client input failures from programmer/configuration
failures. Malformed JSON, empty bodies, extra JSON values, oversized bodies, and
invalid query values remain 4xx request errors. Invalid bind destinations, nil
`Ctx`, nil requests, invalid bind options, and validation rule configuration
errors are treated as server errors because callers must fix code or wiring
rather than asking the client to retry with different input.
