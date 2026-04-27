# Card 0482: Router Print Output Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: router
Owned Files:
- `router/router_test.go`
- `router/router_advanced_test.go`
Depends On: none

Goal:
Consolidate router print-output substring assertions behind a shared test
helper.

Problem:
Router print tests repeat `strings.Contains` checks and ad hoc error messages
for the same output contract.

Scope:
- Add a local helper for asserting printed route output contains expected text.
- Use it in basic and advanced print tests.

Non-goals:
- Do not change router behavior or output format.
- Do not add dependencies.

Files:
- `router/router_test.go`
- `router/router_advanced_test.go`

Tests:
- `go test -race -timeout 60s ./router/...`
- `go test -timeout 20s ./router/...`
- `go vet ./router/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- Print output tests use a shared helper for expected substrings.
- The listed validation commands pass.

Outcome:
