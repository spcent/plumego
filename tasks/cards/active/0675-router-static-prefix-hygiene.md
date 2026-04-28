# Card 0675

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: router
Owned Files: router/static.go, router/static_test.go
Depends On: 0674

Goal:
Make static mount prefix handling canonical for root, trailing-slash, and nil-filesystem cases.

Scope:
- Normalize static mount prefixes by trimming whitespace and trailing slashes while preserving root.
- Register root static mounts as `/*filepath`, not `//*filepath`.
- Ensure trailing-slash prefixes such as `/static/` register the same route shape as `/static`.
- Reject nil `http.FileSystem` values through the public `StaticFS` error path.

Findings:
- `Static("/static/", dir)` currently builds `/static//*filepath`, introducing an empty segment that does not match ordinary `/static/file` requests.
- `Static("/", dir)` currently builds `//*filepath`, which works through trimming side effects but stores a non-canonical pattern.
- `StaticFS(prefix, nil)` registers successfully and can panic on request dispatch.

Non-goals:
- Do not add cache headers, SPA fallback, ETag, precompressed asset, or frontend asset policy.
- Do not change static serving into a frontend capability; keep it a small file-mount primitive.
- Do not change symlink escape behavior beyond preserving existing tests.

Files:
- router/static.go
- router/static_test.go

Tests:
- go test -race -timeout 60s ./router/...
- go test -timeout 20s ./router/...
- go vet ./router/...

Docs Sync:
Not required; this card preserves documented static primitive behavior.

Done Definition:
- `/static`, `/static/`, `static`, and root static prefixes produce deterministic route patterns and serve expected files.
- Nil `StaticFS` registration fails before serving.
- Static path traversal and symlink escape protections remain intact.
