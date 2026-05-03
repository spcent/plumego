# Card 0718

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: router
Owned Files: router/metadata.go, router/registration.go, router/reverse_routing_group_test.go, docs/modules/router/README.md
Depends On:

Goal:
Make named route collisions explicit instead of silently using last registration wins.

Scope:
- Return an AddRoute error when WithRouteName registers a name that already exists.
- Keep existing route metadata behavior for unique names.
- Update tests and docs to describe the collision contract.

Non-goals:
- Adding overwrite options.
- Changing URL generation for unique routes.
- Changing route matching.

Files:
- router/metadata.go
- router/registration.go
- router/reverse_routing_group_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- docs/modules/router/README.md

Done Definition:
- Duplicate named routes return errors at registration time.
- Unique named route behavior remains unchanged.
- Router tests and vet pass.

Outcome:

