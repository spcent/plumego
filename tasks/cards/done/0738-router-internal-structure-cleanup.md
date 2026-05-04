# Card 0738

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/router.go, router/registration.go, router/matcher.go, docs/modules/router/README.md
Depends On: 0737-router-static-file-contract

Goal:
Reduce internal router structure drift by removing stale group parent state and
deduplicating trie child lookup helpers.

Scope:
- Remove the unused `Router.parent` field.
- Share child lookup helpers between registration and matching where practical.
- Preserve trie precedence and route registration behavior.
- Document that Router values are not intended to be copied.

Non-goals:
- Public API additions.
- Replacing the trie data structure.
- Performance threshold changes.

Files:
- router/router.go
- router/registration.go
- router/matcher.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- No stale `parent` field remains.
- Registration and matching use consistent child lookup behavior.
- Router tests, race tests, and vet pass.

Outcome:
- Removed the unused `Router.parent` field.
- Consolidated static, parameter, and wildcard child lookup helpers so
  registration and matching share consistent trie traversal behavior.
- Documented that callers should use the `*Router` returned by `NewRouter` or
  `Group` and should not copy Router values.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
