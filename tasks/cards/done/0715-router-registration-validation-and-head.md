# Card 0715

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: router
Owned Files: router/registration.go, router/dispatch.go, router/negative_test.go, router/router_contract_test.go
Depends On: 0714-router-match-cache-correctness

Goal:
Make invalid route registration fail through returned errors and tighten HEAD fallback writer behavior.

Scope:
- Reject route patterns with empty static path segments instead of allowing insertion-time panics.
- Preserve the public AddRoute error contract for malformed routes.
- Make HEAD fallback body suppression report successful byte counts to handlers.

Non-goals:
- Changing trailing slash normalization.
- Changing public method names.
- Adding middleware or response policy.

Files:
- router/registration.go
- router/dispatch.go
- router/negative_test.go
- router/router_contract_test.go

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required unless registration behavior docs need a new explicit note.

Done Definition:
- `/a//b` registration returns an error and does not panic.
- HEAD fallback suppresses body while `Write` returns `len(p), nil`.
- Router tests and vet pass.

Outcome:
Rejected empty route path segments through AddRoute errors and adjusted HEAD fallback body suppression to report the accepted byte count to handlers.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
