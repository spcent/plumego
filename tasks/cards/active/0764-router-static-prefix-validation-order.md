# Card 0764

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: router
Owned Files: router/static.go, router/static_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0763-router-wildcard-empty-segment-contract

Goal:
Make `Static` and `StaticFS` input validation precedence stable and explicit.

Scope:
- Validate normalized static route prefixes before filesystem resolution or nil
  filesystem validation.
- Preserve lifecycle errors as the first precedence class.
- Add regression coverage for invalid static prefixes with missing directories
  and nil file systems.
- Sync docs for lifecycle, prefix, and filesystem validation order.

Non-goals:
- Changing static serving behavior.
- Changing duplicate route precedence.
- Adding frontend asset policy.

Files:
- router/static.go
- router/static_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Invalid static prefixes fail before filesystem or nil `StaticFS` input work.
- Frozen/uninitialized lifecycle errors still take precedence.
- Router targeted tests, race tests, and vet pass.

Outcome:
