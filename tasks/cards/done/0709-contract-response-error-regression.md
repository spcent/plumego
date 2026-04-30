# Card 0709

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: contract
Owned Files: contract/response.go, contract/errors.go, contract/context.go, contract/*_test.go
Depends On: 0706-stable-root-api-inventory

Goal:
Verify `contract` response, error, and context access behavior stays on the canonical paths.

Scope:
Add or tighten tests for `WriteResponse`, `WriteError`, error builder output, request metadata, and context accessors using existing exported APIs.

Non-goals:
Do not add a second response envelope.
Do not introduce new error-helper families.
Do not move session or observability behavior into `contract`.

Files:
contract/response.go
contract/errors.go
contract/context.go
contract/*_test.go

Tests:
go test -race -timeout 60s ./contract/...
go test -timeout 20s ./contract/...
go run ./internal/checks/dependency-rules

Docs Sync:
None unless behavior changes. If existing docs describe a different response/error shape, record the drift in Outcome.

Done Definition:
Response and error paths have positive and negative regression tests.
Context accessors remain explicit and stable.
Contract package tests and dependency boundary check pass.

Outcome:
Completed.

Changes:

- Added a `WriteResponse` regression test proving the canonical success envelope includes `data`, `meta`, and context-derived `request_id`.
- No runtime code or public API changed.

Validation:

- `go test -race -timeout 60s ./contract/...` passed.
- `go test -timeout 20s ./contract/...` passed.
- `go run ./internal/checks/dependency-rules` passed.

