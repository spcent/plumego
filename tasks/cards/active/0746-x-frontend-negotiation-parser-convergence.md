# Card 0746: x/frontend Negotiation Parser Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/compression.go`
- `x/frontend/negotiation.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
Depends On: 0745

Goal:
Converge the small hand-written `Accept` and `Accept-Encoding` quality parsing
paths into one internal helper.

Scope:
- Extract shared token/media-range quality parsing into an internal helper.
- Keep current `Accept` and `Accept-Encoding` behavior unchanged.
- Add regression tests that cover invalid q values in both paths.
- Keep the helper stdlib-only and package-private.

Non-goals:
- Do not add full HTTP content negotiation beyond current supported behavior.
- Do not add runtime compression.
- Do not change public API.

Files:
- `x/frontend/frontend.go`
- `x/frontend/compression.go`
- `x/frontend/negotiation.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
No docs change required unless behavior wording changes.

Done Definition:
- Quality parsing is implemented in one internal helper.
- Existing negotiation behavior and tests remain passing.
- The listed validation commands pass.
