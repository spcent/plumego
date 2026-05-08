# Card 1158

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: done
Primary Module: router
Owned Files: router/static.go, router/static_test.go, tasks/cards/active/README.md
Depends On: 0753-router-head-comment-sync

Goal:
Remove dead complexity from static prefix normalization.

Scope:
- Collapse duplicate `strings.TrimRight` and empty checks in
  `normalizeStaticPrefix`.
- Keep current prefix behavior for empty, root, relative, absolute, and trailing
  slash inputs.
- Add focused table coverage for normalization edge cases if missing.

Non-goals:
- Changing static route pattern semantics.
- Changing static serving behavior.
- Adding frontend asset policy.

Files:
- router/static.go
- router/static_test.go
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required.

Done Definition:
- Static prefix normalization has no duplicate trim branch.
- Existing and new static prefix tests pass.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Removed the duplicate `strings.TrimRight` and empty-check branch from
  `normalizeStaticPrefix`.
- Added table coverage for empty, root, relative, absolute, repeated root, and
  trailing-slash prefixes.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
