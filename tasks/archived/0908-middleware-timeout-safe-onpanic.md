# Card 0908

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/timeout/timeout.go
- middleware/timeout/timeout_test.go
- docs/modules/middleware/README.md
Depends On:
- 0732-middleware-recovery-panic-log-sanitization

Goal:
Make timeout late-panic reporting safe when the parent request goroutine has
already returned.

Scope:
- Wrap `TimeoutConfig.OnPanic` invocation so callback panics do not escape the
  timeout worker goroutine.
- Preserve pre-timeout panic propagation to outer recovery middleware.
- Document that `OnPanic` is best-effort and must be non-blocking.

Non-goals:
- Do not make timeout wait for late downstream goroutines.
- Do not change the timeout response envelope.
- Do not introduce background worker pools.

Files:
- middleware/timeout/timeout.go
- middleware/timeout/timeout_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/timeout
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Late downstream panic invokes `OnPanic` when configured.
- A panic inside `OnPanic` is recovered internally.
- Targeted and middleware-wide tests pass.

Outcome:
- Wrapped late `TimeoutConfig.OnPanic` invocation with an internal recover guard.
- Added regression coverage for callback panics after a timeout response has
  already returned.
- Documented the hook as best-effort and non-blocking.

Validation:
- `go test -timeout 20s ./middleware/timeout`
- `go test -timeout 20s ./middleware/...`
