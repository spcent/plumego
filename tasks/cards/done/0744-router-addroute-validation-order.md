# Card 0744

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/registration.go, router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0743-router-nil-url-guard

Goal:
Make `AddRoute` validation order and nil-handler errors canonical.

Scope:
- Check router readiness before route input validation.
- Check frozen lifecycle before method and handler validation.
- Remove the duplicate nil-handler check and keep one error message.
- Add focused regression coverage for ready, zero-value, and frozen routers.
- Sync router lifecycle docs if behavior wording changes.

Non-goals:
- Changing route matching behavior.
- Adding new exported error types.
- Changing public route registration APIs.

Files:
- router/registration.go
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required if lifecycle precedence wording changes.

Done Definition:
- `AddRoute` has one nil-handler validation path.
- Uninitialized and frozen lifecycle errors take precedence consistently.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Reordered `AddRoute` validation so uninitialized and frozen lifecycle errors
  take precedence before method and handler input validation.
- Removed the duplicate nil-handler branch and kept one canonical `nil handler`
  error message.
- Added regression coverage for nil/zero routers, frozen routers, ready method
  validation, and nil-handler errors.
- Documented lifecycle error precedence for direct router registration.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
