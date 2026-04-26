# Card 2212: Router Static File Primitive Hardening

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: router
Owned Files:
- `router/static.go`
- `router/static_test.go`
Depends On: 2211

Goal:
Tighten static file primitive path safety without expanding router into
frontend asset policy.

Problem:
Static serving has several local helpers with overlapping path checks. The
relative path cleaner rejects any cleaned path containing `..`, which is
defensive but imprecise, and the directory root check allows missing target
paths before the existence check. The behavior should be covered directly so
future edits do not weaken traversal and symlink protections.

Scope:
- Keep static mounts as GET-only file primitives.
- Make relative path validation explicit and easy to audit.
- Add focused tests for traversal, absolute paths, null bytes, and symlink
  escape rejection.

Non-goals:
- Do not add cache headers, SPA fallback, precompressed assets, custom MIME
  policy, ETags, or frontend policy.
- Do not add dependencies.
- Do not change `Static` or `StaticFS` public signatures.

Files:
- `router/static.go`
- `router/static_test.go`

Tests:
- `go test -race -timeout 60s ./router/...`
- `go test -timeout 20s ./router/...`
- `go vet ./router/...`

Docs Sync:
No docs change required; this preserves existing static primitive scope.

Done Definition:
- Unsafe paths and symlink escapes return 404.
- Valid nested files still serve correctly.
- StaticFS behavior remains intact.

Outcome:
- Replaced broad substring traversal rejection with path-component traversal
  detection.
- Added coverage proving safe filenames containing `..` are allowed.
- Added symlink escape coverage for local directory static mounts.
