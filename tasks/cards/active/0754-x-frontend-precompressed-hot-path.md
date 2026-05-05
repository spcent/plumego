# Card 0754: x/frontend Precompressed Hot Path

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/compression.go`
- `x/frontend/compression_test.go`
- `x/frontend/README.md`
Depends On: 0753

Goal:
Reduce redundant file opens when a precompressed variant can satisfy a request.

Scope:
- Prefer accepted precompressed variants before opening the original asset when
  directory-backed variant metadata proves a candidate exists.
- Preserve original-file existence checks before serving orphan variants.
- Keep custom filesystem lazy behavior unchanged unless correctness requires a
  fallback.
- Add regression coverage for served variants and orphan rejection.

Non-goals:
- Do not add runtime compression.
- Do not change response headers or negotiation semantics.
- Do not add public API.

Files:
- `x/frontend/frontend.go`
- `x/frontend/compression.go`
- `x/frontend/compression_test.go`
- `x/frontend/README.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go test -race -timeout 60s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Update docs only if the internal serving order is mentioned.

Done Definition:
- Directory-backed precompressed hits avoid unnecessary original file opens.
- Orphan variants are still not served.
- The listed validation commands pass.
