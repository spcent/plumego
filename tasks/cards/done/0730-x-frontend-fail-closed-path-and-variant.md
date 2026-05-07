# Card 0730: x/frontend Fail-Closed Path and Variant Semantics

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
Depends On: 0729

Goal:
Ensure unsafe request paths fail closed and precompressed responses are only
served as variants of existing source assets.

Scope:
- Return 404 for invalid request paths even when SPA fallback is enabled.
- Require the original requested asset to exist before serving `.br` or `.gz`
  variants.
- Normalize prefix validation using path segments instead of substring checks.
- Add regression coverage for dotted prefix segments, invalid fallback paths,
  and orphan precompressed variants.

Non-goals:
- Do not change router static primitives.
- Do not add runtime compression.
- Do not change public option names.

Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
No docs change required unless behavior text changes.

Done Definition:
- Traversal/null/backslash paths never return SPA index responses.
- `/app..v2` style prefixes are allowed.
- Orphan `.br`/`.gz` files are not served for missing originals.
- The listed validation commands pass.

Outcome:
- Invalid request paths now fail closed with 404 instead of returning the SPA
  fallback index.
- Precompressed variants are only considered after the original requested asset
  exists and is not a directory.
- Prefix validation now rejects only unsafe path segments and allows dotted
  prefix names such as `/app..v2`.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
