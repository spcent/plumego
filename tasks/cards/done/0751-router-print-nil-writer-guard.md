# Card 0751

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/metadata.go, router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0750-router-lifecycle-test-contract

Goal:
Harden `Router.Print` so nil writers do not panic on the public surface.

Scope:
- Guard nil `io.Writer` before writing the routes header.
- Add regression coverage for nil/zero routers and ready routers with nil
  writer.
- Document the defensive no-op behavior if needed.

Non-goals:
- Changing normal route print output.
- Changing the public `Print` signature.
- Adding structured route printing.

Files:
- router/metadata.go
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required if public behavior is documented.

Done Definition:
- `Print(nil)` is a no-op and does not panic.
- Existing print output tests keep passing.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Added an early nil writer guard to `Router.Print`.
- Added regression coverage for `Print(nil)` on nil, zero-value, and ready
  routers.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
