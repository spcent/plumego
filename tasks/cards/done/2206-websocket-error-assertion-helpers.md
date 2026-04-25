# Card 2206: Websocket Error Assertion Helpers

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/websocket_test.go`
- `x/websocket/websocket_extended_test.go`
Depends On: none

Goal:
Consolidate repeated websocket error assertion patterns.

Problem:
Websocket tests repeat nil checks, `errors.Is` checks, and fallback substring
checks for closed, timeout, queue-full, and unknown-behavior errors. The
patterns are equivalent but scattered across two test files.

Scope:
- Add package-level test helpers for required error substrings and
  `errors.Is`-or-substring fallbacks.
- Use them in websocket constructor and connection send behavior tests.

Non-goals:
- Do not change websocket runtime behavior or public APIs.
- Do not add dependencies.

Files:
- `x/websocket/websocket_test.go`
- `x/websocket/websocket_extended_test.go`

Tests:
- `go test -race -timeout 60s ./x/websocket/...`
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- Repeated websocket error assertions use named helpers.
- The listed validation commands pass.

Outcome:
- Added `assertErrorContains` and `assertErrorIsOrContains` for websocket tests.
- Validation passed for websocket race tests, normal tests, and vet.
