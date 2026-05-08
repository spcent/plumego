# Card 0957

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/static.go, router/static_test.go, docs/modules/router/README.md
Depends On: 0736-router-request-path-canonicalization

Goal:
Make `Static` and `StaticFS` consistently serve regular files only.

Scope:
- Reject directory requests for both local directory and custom filesystem
  static mounts.
- Keep nonexistent and unsafe paths as 404.
- Add tests for directory requests under `Static` and `StaticFS`.
- Document the regular-file primitive contract.

Non-goals:
- Directory listings.
- Index file fallback.
- SPA fallback or frontend asset policy.

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
- Static directory requests return 404 consistently.
- Static file requests still work.
- Router tests, race tests, and vet pass.

Outcome:
- Static and StaticFS now return 404 for directory requests and continue
  serving regular files.
- Removed the redundant local static existence helper in favor of direct
  `os.Stat` plus directory rejection.
- Added local directory and `http.FileSystem` directory request coverage.
- Updated router docs for the regular-file-only static primitive contract.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
