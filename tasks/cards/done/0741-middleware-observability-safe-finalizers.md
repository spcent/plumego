# Card 0741

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/internal/observability/helpers.go
- middleware/accesslog/accesslog.go
- middleware/httpmetrics/http_metrics.go
- middleware/tracing/tracing.go
- middleware/accesslog/accesslog_test.go
- middleware/httpmetrics/http_metrics_test.go
- middleware/tracing/tracing_test.go
Depends On:
- 0740-middleware-coalesce-safe-hooks

Goal:
Prevent observability finalizers from replacing downstream handler panics.

Scope:
- Add a shared safe finalizer helper for observability middleware defer paths.
- Use it in accesslog, httpmetrics, and tracing.
- Preserve original downstream panic values while recovering finalizer panics.

Non-goals:
- Do not change observer/tracer/logger public interfaces.
- Do not add dependencies.
- Do not redesign observability composition.

Files:
- middleware/internal/observability/helpers.go
- middleware/accesslog/accesslog.go
- middleware/httpmetrics/http_metrics.go
- middleware/tracing/tracing.go
- middleware/accesslog/accesslog_test.go
- middleware/httpmetrics/http_metrics_test.go
- middleware/tracing/tracing_test.go

Tests:
- go test -timeout 20s ./middleware/accesslog ./middleware/httpmetrics ./middleware/tracing
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Finalizer panics are recovered internally.
- A downstream panic remains the panic observed by outer recovery.
- Targeted and middleware-wide tests pass.

Outcome:
- Added a shared observability finalizer helper that recovers finalizer panics
  and re-panics any original downstream panic.
- Accesslog, httpmetrics, and tracing now use the helper in their defer paths.
- Added regression coverage for logger, metrics collector, and span finalizer
  panics during downstream panic unwinding.

Validation:
- `go test -timeout 20s ./middleware/accesslog ./middleware/httpmetrics ./middleware/tracing`
- `go test -timeout 20s ./middleware/...`
