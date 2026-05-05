# Card 0758

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: router
Owned Files: router/path.go, router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0757-router-static-containment-contract

Goal:
Make request trailing slash normalization consistent for one or more trailing
slashes.

Scope:
- Normalize all trailing slashes on non-root request paths.
- Preserve rejection of internal empty segments such as `/users//123`.
- Add regression coverage for multiple trailing slashes.
- Sync docs for request path normalization.

Non-goals:
- Changing route registration normalization.
- Changing encoded slash behavior.
- Adding redirect behavior.

Files:
- router/path.go
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- `/users/123/` and `/users/123//` resolve consistently.
- Internal empty segments remain non-matches.
- Router targeted tests, race tests, and vet pass.

Outcome:
