# contract

## Purpose

`contract` defines Plumego's HTTP transport contracts: structured API errors,
success response envelopes, request metadata context helpers, request ids, and
trace metadata carriers.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public surface is intentionally small and stdlib-compatible

## Use this module when

- standardizing success or error response shape
- writing HTTP transport response helpers
- carrying router-owned request metadata needed by handlers or middleware
- carrying request id or trace metadata through `context.Context`

## Do not use this module for

- request body binding
- query binding
- validation frameworks or `validate` tag processing
- protocol gateway families
- business DTO ownership
- route matching
- auth identity contracts
- middleware observability policy or field redaction

## First files to read

- `contract/module.yaml`
- `contract/errors.go`
- `contract/response.go`
- `contract/context_core.go`
- `contract/request_id.go`
- `contract/trace.go`

## Public entrypoints

- `HeaderContentType`
- `ContentTypeJSON`
- `NewErrorBuilder`
- `ErrorBuilder`
- `ErrorCategory`
- `ErrorType`
- `ErrorTypeMeta`
- `ErrorType.Meta`
- `APIError`
- `ErrorResponse`
- `Code*` transport error constants
- `WriteError`
- `WriteResponse`
- `RequestContext`
- `WithRequestContext`
- `RequestContextFromContext`
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
- `TraceContext`
- `TraceContext.IsSampled`
- `TraceContext.HasTraceID`
- `TraceContext.HasSpanID`
- `TraceContext.Valid`
- `ErrHandlerNil`
- `ErrResponseWriterNil`

## Surface Rules

`contract` owns one success response path and one structured error write path:

- success responses go through `WriteResponse`
- errors go through `NewErrorBuilder` and `WriteError`
- handlers decode JSON directly with `json.NewDecoder(r.Body).Decode(&dst)`
- handlers read query values directly from `r.URL.Query()`
- validation belongs in the module that owns the request DTO or business rule

`contract` intentionally does not expose a framework-style request bag, JSON
binder, query binder, validation tag framework, or bind-error translation helper.

## Error Taxonomy

`ErrorType.Meta()` is the canonical taxonomy lookup. It owns the default
`Status`, `Category`, and `Code` for every public error type. Builder and writer
normalization must converge back to this table for status and category.
Coarse status-to-category fallback is internal normalization only; callers
should choose a specific `ErrorType` whenever they know the failure shape.

## Context Metadata

`RequestContext` is router metadata only:

- route params
- matched route pattern
- route name

Use `WithRequestContext` and `RequestContextFromContext`; do not write raw
context values with exported keys. Stored maps are defensively copied.

`TraceContext` is a carrier only. Full tracing runtime behavior, propagation
policy, collectors, samplers, and exporters belong in `x/observability`.

## Validation

Module-owned handlers should keep transport validation explicit:

```go
var req createRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    _ = contract.WriteError(w, r, contract.NewErrorBuilder().
        Type(contract.TypeValidation).
        Code(contract.CodeInvalidJSON).
        Message("invalid request body").
        Build())
    return
}
if strings.TrimSpace(req.Name) == "" {
    _ = contract.WriteError(w, r, contract.NewErrorBuilder().
        Type(contract.TypeRequired).
        Code(contract.CodeRequired).
        Message("name is required").
        Detail("field", "name").
        Build())
    return
}
```

Complex validation, domain policy, and service-specific error typing should
remain in the owning stable module, extension module, or reference application.
