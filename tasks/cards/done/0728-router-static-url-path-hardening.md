# Card 0728

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/static.go, router/static_test.go, docs/modules/router/README.md
Depends On: 0727-router-lifecycle-zero-value-guards

Goal:
Use URL slash semantics for static file path extraction before converting to
local filesystem paths.

Scope:
- Use slash-based path cleaning for request/static FS paths.
- Reject parent traversal and backslash traversal before local path conversion.
- Convert cleaned slash paths to local paths only for directory serving.
- Add static regression tests for dot/backslash traversal behavior.

Non-goals:
- Adding frontend cache policy.
- Adding SPA fallback.
- Changing StaticFS registration shape.

Files:
- router/static.go
- router/static_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Static path handling uses URL path semantics consistently.
- Traversal-style inputs remain fail-closed.
- Router tests, race tests, and vet pass.

Outcome:
- Static request paths now reject null bytes, backslashes, absolute paths, and
  parent traversal before slash-based URL path cleaning.
- Directory serving converts cleaned slash paths to platform paths only at the
  local filesystem boundary.
- Added regression coverage for nested traversal, backslash traversal, and dot
  segment cleanup.
- Updated router docs for static URL path semantics.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
