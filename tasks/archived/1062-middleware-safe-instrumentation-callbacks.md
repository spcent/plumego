# Card 1062

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/internal/observability/helpers.go
- middleware/recovery/recover.go
- middleware/bodylimit/body_limit.go
- middleware/ratelimit/abuse_guard.go
- middleware/tracing/tracing.go

Goal:
Ensure logging, observer, and tracing callbacks cannot change canonical
middleware response semantics.

Scope:
- Make recovery logging best-effort so logger panic cannot prevent structured
  recovery responses.
- Make tracer start best-effort so tracer panic cannot prevent downstream
  handlers from running.
- Make bodylimit and ratelimit logger calls best-effort.
- Add focused tests for each panic-isolation path.

Non-goals:
- Do not change logger, metrics, or tracing public interfaces.
- Do not add exporter or backend observability wiring to stable middleware.
- Do not change response payload shapes except preserving existing canonical
  responses when callbacks fail.

Files:
- middleware/internal/observability/helpers.go
- middleware/internal/observability/helpers_test.go
- middleware/recovery/recover.go
- middleware/recovery/recover_test.go
- middleware/bodylimit/body_limit.go
- middleware/bodylimit/body_limit_test.go
- middleware/ratelimit/abuse_guard.go
- middleware/ratelimit/abuse_guard_test.go
- middleware/tracing/tracing_test.go
- middleware/accesslog/accesslog_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/internal/observability ./middleware/recovery ./middleware/bodylimit ./middleware/ratelimit ./middleware/tracing ./middleware/accesslog
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Callback/logger/tracer panics are recovered internally.
- Canonical error responses are still written when callback/logging fails.
- Middleware-wide tests pass.

Outcome:
- Made trace start best-effort in shared observability helpers; tracer `Start`
  panics now skip tracing for that request and still run the downstream handler.
- Wrapped recovery, bodylimit, and ratelimit logging in safe finalizers so
  logger panics cannot block canonical 500/413/429 responses.
- Added regression tests for tracer start panic and logger panic isolation in
  recovery, bodylimit, ratelimit, tracing, accesslog, and shared observability.
- Updated middleware docs to state callback/logging failures are best-effort.

Validation:
- go test -timeout 20s ./middleware/internal/observability ./middleware/recovery ./middleware/bodylimit ./middleware/ratelimit ./middleware/tracing ./middleware/accesslog
- go test -timeout 20s ./middleware/...
