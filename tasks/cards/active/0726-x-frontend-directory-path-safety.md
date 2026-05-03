# Card 0726: x/frontend Directory Path Safety

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
Depends On: none

Goal:
Harden frontend directory serving and request path validation before stable
promotion.

Scope:
- Prevent local-directory symlink escapes when serving via `RegisterFromDir`.
- Replace substring-based traversal checks with segment-aware path validation.
- Reject null bytes, absolute paths, parent traversal segments, and backslash
  traversal forms consistently.
- Add targeted regression coverage for safe dotted filenames and unsafe paths.

Non-goals:
- Do not change router static primitives.
- Do not add dependencies.
- Do not change frontend mount registration semantics.

Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
No docs change required unless public behavior text changes.

Done Definition:
- Directory-backed frontend serving cannot follow symlinks outside the mounted
  root.
- Legal filenames containing `..` as ordinary characters still serve.
- Unsafe request paths are rejected or fall back without exposing files.
- The listed validation commands pass.

Outcome:

