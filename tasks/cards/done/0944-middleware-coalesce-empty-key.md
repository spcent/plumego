# Card 0944

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
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
- Blank or all-whitespace custom coalesce keys now fail open and bypass
  in-flight coalescing state.
- Added regression coverage proving blank-key requests are not coalesced and do
  not receive `X-Coalesced`.
- Documented the blank-key pass-through contract.

Validation:
- `go test -timeout 20s ./middleware/coalesce`
- `go test -timeout 20s ./middleware/...`
