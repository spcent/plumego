# Card 0754: x/frontend Precompressed Hot Path

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: done
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

Outcome:
- Directory-backed planned precompressed variants are now attempted before
  opening the original asset. The optimized path verifies the current original
  path still exists inside the mounted directory before serving the variant.
- Custom filesystem lazy probing remains unchanged because those filesystems do
  not expose construction-time directory metadata.
- Added regression coverage that removes an original asset after planning and
  confirms the precompressed variant is not served.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go vet ./x/frontend/...`
