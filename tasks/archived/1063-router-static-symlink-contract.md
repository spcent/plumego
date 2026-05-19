# Card 1063

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/static.go, router/static_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0745-router-static-lifecycle-order

Goal:
Clarify and tighten static local-file symlink escape handling for the stable
primitive contract.

Scope:
- Avoid the current check-then-serve reopen path for local directory serving.
- Serve already-opened regular files after root containment validation where
  possible.
- Keep directory requests and symlink escapes returning 404.
- Add or adjust regression coverage for symlink escape and normal file serving.
- Sync docs to describe the implemented static primitive behavior.

Non-goals:
- Adding index files, directory listing, SPA fallback, cache headers, or ETags.
- Changing `StaticFS` behavior beyond shared helper reuse if needed.
- Introducing non-stdlib dependencies.

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
- Local static serving does not reopen the path after containment validation.
- Existing static file and symlink escape regressions pass.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Local static serving now resolves the requested path, verifies the resolved
  path stays inside the static root, opens that resolved path, and serves the
  already opened file handle.
- Removed the previous containment-check plus `http.ServeFile` reopen path.
- Preserved 404 behavior for symlink escapes and directory requests.
- Added coverage for symlinks that resolve inside the static root.
- Documented the local static symlink and open-file behavior.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
