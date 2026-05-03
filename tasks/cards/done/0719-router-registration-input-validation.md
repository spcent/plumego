# Card 0719

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/registration.go, router/router_contract_test.go, docs/modules/router/README.md
Depends On: 0718-router-named-route-collision-contract

Goal:
Tighten AddRoute input validation for method names and route parameter names.

Scope:
- Reject empty or whitespace-padded methods.
- Reject parameter and wildcard names outside the stable identifier shape.
- Keep custom non-standard HTTP methods possible when they are clean tokens.

Non-goals:
- Removing ANY support.
- Changing path normalization.
- Adding route parameter value validation.

Files:
- router/registration.go
- router/router_contract_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- docs/modules/router/README.md

Done Definition:
- Invalid methods and parameter names return AddRoute errors.
- Existing valid router tests still pass.

Outcome:
Added AddRoute validation for empty or whitespace-containing methods and ASCII identifier route parameter names. Custom clean method tokens remain supported.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
