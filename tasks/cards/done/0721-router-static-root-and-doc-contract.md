# Card 0721

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/static.go, router/static_test.go, docs/modules/router/README.md
Depends On: 0720-router-head-body-writer-compatibility

Goal:
Resolve Static local roots at registration time and align StaticFS source comments with embedded directory usage.

Scope:
- Return an error from Static when the local root cannot be resolved.
- Reuse a canonical root path per handler instead of resolving it on every request.
- Update StaticFS doc comments and focused tests.

Non-goals:
- Adding frontend cache/fallback policy.
- Changing StaticFS semantics.
- Moving behavior to x/frontend.

Files:
- router/static.go
- router/static_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- docs/modules/router/README.md

Done Definition:
- Static fails fast for missing or invalid roots.
- Static still blocks symlink escape.
- Router tests and vet pass.

Outcome:
Changed Static to resolve and validate local roots during registration, reusing the canonical root in the handler. Updated StaticFS source comments and docs to show fs.Sub for embedded directories.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
