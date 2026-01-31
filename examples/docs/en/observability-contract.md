# Observability Contract

This document defines the default observability conventions for Plumego.
It does not require a tracing SDK, but standardizes context keys, headers,
and log field names for consistent troubleshooting.

## Request ID / Trace ID
- Preferred header: `X-Request-ID` (case-insensitive).
- If present, the value should be propagated into context.
- If missing, middleware should generate one and inject it into:
  - Response header `X-Request-ID`
  - Context key `contract.TraceIDKey{}`
  - `middleware.RequestID()` provides this behavior without requiring a tracer.

## Trace Context
- Use `contract.TraceContext` when available.
- Read with `contract.TraceContextFromContext(ctx)`.
- Fallback to `contract.TraceIDFromContext(ctx)` if no trace context exists.

## Context Field Conventions
- Read-only identifiers: `trace_id`, `span_id`, `request_id`
- Request metadata: method, path, status, duration, bytes
- Route metadata (when available): `route` (pattern) and `route_name`
- Security-sensitive fields should never include secrets or raw tokens.

## Log Field Names (Recommended)
Minimal fields for production troubleshooting:
- `trace_id`
- `method`
- `path`
- `route`
- `status`
- `duration_ms`
- `bytes`

Additional fields when available:
- `span_id`
- `request_id`
- `client_ip`
- `user_agent`
- `route_name`
- `error`
- `error_code`
- `error_type`
- `error_category`

## Error Mapping
- If using `contract.APIError`, map fields to log keys:
  - `error_code` from `APIError.Code`
  - `error_category` from `APIError.Category`
  - `error_type` from `APIError.Details["type"]` when present
