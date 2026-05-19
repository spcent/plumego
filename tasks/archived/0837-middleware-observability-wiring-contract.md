# Card 0837

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- docs/modules/middleware/README.md
- reference/standard-service/internal/app/app.go
- middleware/accesslog/accesslog.go
- middleware/accesslog/accesslog_test.go
Depends On:
- 0726-middleware-timeout-late-panic-reporting

Goal:
Make the stable observability wiring pattern explicit enough to avoid duplicate
HTTP observations or duplicate tracing spans.

Scope:
- Define one recommended production mode: standalone `httpmetrics` and
  `tracing` middleware, with `accesslog` used for logging only.
- Keep `accesslog` observer/tracer parameters as compatibility wiring, but
  document that callers should not also stack the standalone middleware for the
  same concern.
- Align reference-service wiring or examples with the recommended pattern.
- Add or update tests only if behavior is clarified in code comments or docs.

Non-goals:
- Do not remove public accesslog parameters.
- Do not add runtime duplicate-detection globals.
- Do not move exporter, sampler, or backend setup into stable middleware.

Files:
- docs/modules/middleware/README.md
- reference/standard-service/internal/app/app.go
- middleware/accesslog/accesslog.go
- middleware/accesslog/accesslog_test.go

Tests:
- go test -timeout 20s ./middleware/accesslog ./middleware/httpmetrics ./middleware/tracing
- go test -timeout 20s ./reference/standard-service/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Docs name the canonical composition and compatibility-only accesslog wiring.
- Reference wiring does not encourage duplicate observer/tracer attachment.
- Targeted observability and reference tests pass.

Outcome:
- Documented the recommended production observability composition:
  standalone `httpmetrics` and `tracing`, with `accesslog` logging-only.
- Clarified that accesslog observer/tracer parameters are compatibility wiring
  and should not be combined with standalone middleware for the same signal.
- Added the logging-only reference wiring comment to standard-service.

Validation:
- `go test -timeout 20s ./middleware/accesslog ./middleware/httpmetrics ./middleware/tracing`
- `go test -timeout 20s ./reference/standard-service/...`
