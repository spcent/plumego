# Card 0485: Middleware Debug Content-Type Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: middleware/debug
Owned Files:
- `middleware/debug/debug_errors_test.go`
Depends On: none

Goal:
Use a local helper for debug error content-type assertions.

Problem:
The debug error test checks JSON content type with an inline substring
assertion. A helper keeps the expected transport shape named and reusable.

Scope:
- Add a small test helper for JSON content-type assertions.
- Use it in the not-found debug error test.

Non-goals:
- Do not change middleware behavior or response shape.
- Do not add dependencies.

Files:
- `middleware/debug/debug_errors_test.go`

Tests:
- `go test -race -timeout 60s ./middleware/debug/...`
- `go test -timeout 20s ./middleware/debug/...`
- `go vet ./middleware/debug/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- Debug error content-type assertion uses the local helper.
- The listed validation commands pass.

Outcome:
