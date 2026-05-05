# Card 0736

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md
Depends On:
- 0735-middleware-gzip-flush-boundary

Goal:
Prevent empty custom coalesce keys from merging unrelated requests.

Scope:
- Treat a blank key from `KeyFunc` as fail-open pass-through.
- Trim only for the emptiness check; do not otherwise rewrite caller-provided
  keys.
- Add tests for blank key pass-through behavior.

Non-goals:
- Do not change default key generation.
- Do not add a new error response for blank keys.
- Do not redesign coalescing hooks.

Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/coalesce
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Blank keys do not create or join an in-flight coalescing slot.
- Normal keyed coalescing behavior remains unchanged.
- Targeted and middleware-wide tests pass.

Outcome:

