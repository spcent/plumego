# Card 0726

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
- middleware/timeout/timeout.go
- middleware/timeout/timeout_test.go
- docs/modules/middleware/README.md
Depends On:
- 0725-middleware-cors-wildcard-header-normalization

Goal:
Make timeout handler panics deterministic when they happen after the timeout
response has already been emitted.

Scope:
- Keep pre-timeout downstream panics on the parent goroutine so outer recovery
  middleware can handle them.
- Add an explicit post-timeout panic reporting path because outer recovery can no
  longer safely rewrite a completed timeout response.
- Document the distinction between recoverable pre-timeout panic and observable
  post-timeout panic.
- Add regression coverage for post-timeout panic reporting.

Non-goals:
- Do not block timeout responses waiting for ignored downstream goroutines.
- Do not crash the process for post-timeout panics.
- Do not add logging globals or package-level observers.

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
- Pre-timeout panics still re-panic to outer recovery.
- Post-timeout panics are observable through explicit configuration.
- Timeout package and middleware-wide tests pass.

Outcome:
