# Card 2213: Router Test Duplication Cleanup

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: done
Primary Module: router
Owned Files:
- `router/test_helpers_test.go`
- `router/router_contract_test.go`
- `router/router_conflict_test.go`
- `router/static_test.go`
Depends On: 2212

Goal:
Reduce repeated router test boilerplate while preserving behavior coverage.

Problem:
Router tests repeat route registration checks, request execution, status
assertions, and body trimming in many files. This makes behavior tests noisier
and increases the chance of inconsistent failure messages.

Scope:
- Add small local test helpers for common request/status/body assertions.
- Apply helpers to the highest-repetition router contract/conflict/static tests.
- Keep tests explicit where the setup itself is the behavior under test.

Non-goals:
- Do not change production code.
- Do not hide route registration details needed for readability.
- Do not convert table tests just for style.

Files:
- `router/test_helpers_test.go`
- `router/router_contract_test.go`
- `router/router_conflict_test.go`
- `router/static_test.go`

Tests:
- `go test -race -timeout 60s ./router/...`
- `go test -timeout 20s ./router/...`
- `go vet ./router/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- Common router test assertions use shared helpers where it improves clarity.
- Router behavior tests keep their existing coverage.
- The listed validation commands pass.

Outcome:
- Added shared router test helpers for request execution, status, body, trimmed
  body, and header assertions.
- Applied the helpers to high-repetition contract, conflict, and static tests.
- Kept custom setup in place where request context or host mutation is the
  behavior under test.
