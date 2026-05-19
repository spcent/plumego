# Card 1256

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/matcher.go, router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0762-router-method-any-request-contract

Goal:
Make internal empty path segments invalid for wildcard routes too.

Scope:
- Stop wildcard matching from swallowing empty internal path segments.
- Preserve normal wildcard matching for non-empty remaining paths.
- Add regression coverage for `/files//a` against `/files/*path`.
- Sync docs if wording needs clarification.

Non-goals:
- Changing encoded slash behavior.
- Changing trailing slash normalization.
- Changing static mount path cleaning beyond matching behavior.

Files:
- router/matcher.go
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required if wildcard wording changes.

Done Definition:
- Internal double slash paths do not match wildcard routes.
- Existing wildcard route behavior for normal paths remains intact.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Changed wildcard matching so internal empty path segments do not match
  wildcard routes.
- Added regression coverage for `/files//a` against `/files/*path`.
- Preserved normal wildcard capture for non-empty path remainders.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
