# Card 0203

Priority: P0
State: done
Primary Module: contract
Owned Files:
- `contract/trace.go`
- `contract/observability_policy.go`
- `middleware/requestid/request_id.go`
- `log/traceid.go`
- `docs/modules/contract/README.md`
Depends On:

Goal:
- Re-establish one canonical request/trace identity contract for the stable layer so `contract`, `middleware/requestid`, and `log` stop disagreeing on header names, ID format, and generation rules.
- Rebuild this area for one clear end state, not a compatibility-preserving transition.

Problem:
- `contract.TraceID` currently validates 32-char hex IDs, but `middleware/requestid` defaults to `log.NewTraceID()`, which generates a 12-char Base62 ID and forces `contract.WithTraceIDString` onto its warning-and-store-raw fallback path.
- `contract.DefaultObservabilityPolicy()` treats request ID and legacy trace ID as one fallback chain, while the package surface still uses both “request id” and “trace id” terminology interchangeably.
- This leaves the stable observability contract non-canonical: validation says one thing, the default generator emits another thing, and middleware silently papers over the mismatch.

Scope:
- Choose one stable-layer identity contract for request correlation and make the naming explicit in `contract`.
- Align the default generator and header resolution path with that contract instead of relying on invalid-value fallback storage.
- Keep the request ID policy grep-friendly so downstream middleware can consume one stable source of truth.
- Document the canonical contract in the module primer once the code path is singular again.
- Delete the non-canonical path in the same change instead of keeping legacy parsing, aliases, or fallback terminology alive.

Non-goals:
- Do not move full tracing infrastructure into the stable layer.
- Do not add exporter or collector wiring here.
- Do not broaden `contract` into an observability feature catalog.
- Do not preserve backward-compatible dual ID formats or dual naming schemes.

Files:
- `contract/trace.go`
- `contract/observability_policy.go`
- `middleware/requestid/request_id.go`
- `log/traceid.go`
- `docs/modules/contract/README.md`

Tests:
- `go test -timeout 20s ./contract ./log ./middleware/requestid`
- `go test -race -timeout 60s ./contract ./log ./middleware/requestid`
- `go vet ./contract ./log ./middleware/requestid`

Docs Sync:
- Sync the stable request-correlation contract with the `contract` module primer and any inline comments that still blur request ID vs trace ID semantics.

Done Definition:
- The default request-correlation generator emits IDs that satisfy the stable contract instead of intentionally falling through validation warnings.
- `contract` exposes one explicit header-resolution rule for stable request correlation.
- `middleware/requestid` no longer carries local defaults that drift from `contract`.
- The `contract` module docs describe the same canonical identity contract the code enforces.
- Old non-canonical request/trace ID paths are removed in the same PR; no transitional wrappers or “legacy for compatibility” branches remain.

Outcome:
- Completed.
- Unified the stable request-correlation contract on `request_id`, with `X-Request-ID` as the canonical transport header and `request_id` as the canonical log/response field.
- Removed the previous trace-id/request-id mismatch between stable contracts, middleware, and logging/request-correlation helpers.
