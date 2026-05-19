# Card 1337

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- x/observability/otel.go
- x/observability/tracer/tracer.go
- middleware/internal/observability/helpers.go
- middleware/accesslog/accesslog.go
- docs/modules/contract/README.md
Depends On:
- 0773

Goal:
Audit external `TraceContextFromContext` usage and require validity checks where a complete propagation context is needed.

Scope:
- Update propagation-oriented readers to check `TraceContext.Valid()` before using trace/span pairs.
- Keep logging-only span id extraction narrow and intentional.
- Document the usage distinction between propagation and diagnostic logging.

Non-goals:
- Do not add propagation policy to `contract`.
- Do not reject invalid carriers in `WithTraceContext`.
- Do not change trace id or span id formats.

Files:
- x/observability/otel.go
- x/observability/tracer/tracer.go
- middleware/internal/observability/helpers.go
- middleware/accesslog/accesslog.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./x/observability/... ./middleware/...
- go test -timeout 20s ./contract/...
- go vet ./x/observability/... ./middleware/... ./contract/...

Docs Sync:
- Update TraceContext caller guidance.

Done Definition:
- Propagation readers do not promote invalid/span-only carriers into trusted trace contexts.
- Logging-only usage remains documented.
- Target checks pass.

Outcome:
- Updated OpenTelemetry parent trace/span extraction to require `TraceContext.Valid()` before using context carrier data for propagation.
- Updated the tracer active-span lookup to ignore carriers without a valid span id.
- Updated middleware observability fallback extraction to require a complete valid trace context, while keeping access-log span extraction diagnostic-only via `HasSpanID()`.
- Documented the propagation-vs-diagnostic distinction for TraceContext callers.

Validation:
- go test -timeout 20s ./x/observability/... ./middleware/...
- go vet ./x/observability/... ./middleware/... ./contract/...
- go test -timeout 20s ./contract/... (standalone rerun after a parallel cold-scan timeout)
