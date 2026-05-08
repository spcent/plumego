# Card 0755

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files:
- middleware/accesslog/accesslog.go
- middleware/accesslog/accesslog_test.go
- middleware/httpmetrics/http_metrics.go
- middleware/httpmetrics/http_metrics_test.go
- middleware/tracing/tracing.go
Depends On:
- 0716-middleware-gzip-large-response-contract

Goal:
Make accesslog, httpmetrics, and tracing complete their post-processing consistently on downstream panics.

Scope:
- Use defer-based post-processing where needed so observability records are finalized before panic propagation.
- Preserve recovery ownership outside observability middleware.
- Add focused panic-path tests for logging, metrics, and tracing behavior.

Non-goals:
- Do not make observability recover panics.
- Do not merge accesslog, metrics, and tracing APIs.
- Do not add new observer/tracer abstractions.

Files:
- middleware/accesslog/accesslog.go
- middleware/accesslog/accesslog_test.go
- middleware/httpmetrics/http_metrics.go
- middleware/httpmetrics/http_metrics_test.go
- middleware/tracing/tracing.go
- middleware/tracing/tracing_test.go

Tests:
- go test ./middleware/accesslog ./middleware/httpmetrics ./middleware/tracing
- go test ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md only if ordering guidance changes.

Done Definition:
- Panic requests produce deterministic observability records according to middleware order.
- Panics still propagate to recovery or the server.
- Focused package tests and middleware-wide tests pass.

Outcome:
- Changed accesslog, httpmetrics, and tracing post-processing to defer-based completion.
- Preserved panic propagation; observability middleware still does not recover.
- Added panic-path tests for access logs, HTTP metrics, and tracing spans.
- Validation:
  - `go test ./middleware/accesslog ./middleware/httpmetrics ./middleware/tracing`
  - `go test ./middleware/...`
