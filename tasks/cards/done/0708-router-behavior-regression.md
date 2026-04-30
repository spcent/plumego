# Card 0708

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: router
Owned Files: router/router.go, router/group.go, router/reverse.go, router/*_test.go
Depends On: 0706-stable-root-api-inventory

Goal:
Complete high-risk router regression coverage for static routes, params, groups, and reverse routing.

Scope:
Add or tighten tests around route precedence, param extraction, group prefix behavior, middleware inheritance, and reverse-route lookup using existing APIs.

Non-goals:
Do not add new route helper APIs.
Do not change `core` route registration.
Do not introduce app-level routing conventions.

Files:
router/router.go
router/group.go
router/reverse.go
router/*_test.go

Tests:
go test -race -timeout 60s ./router/...
go test -timeout 20s ./router/...
go run ./internal/checks/dependency-rules

Docs Sync:
None unless tests reveal documented behavior drift.

Done Definition:
Router tests cover static-vs-param precedence, params, groups, and reverse routing.
No new public API is added.
Router package tests and dependency boundary check pass.

Outcome:
Completed.

Changes:

- Added a router regression test proving a static route registered after a parameter route still takes precedence for the same segment.
- Verified the parameter route remains available for non-static matches.
- No runtime code or public API changed.

Validation:

- `go test -race -timeout 60s ./router/...` passed.
- `go test -timeout 20s ./router/...` passed.
- `go run ./internal/checks/dependency-rules` passed.

