# Card 0843: x/frontend HTTP Cache and Compression Semantics

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
Depends On: 0726

Goal:
Make precompressed asset negotiation, cache variance, and method responses
stable-grade HTTP behavior.

Scope:
- Respect `Accept-Encoding` quality factors, including `q=0`.
- Ensure `Vary: Accept-Encoding` is emitted when a URL has precompressed
  variants and response selection can vary by request header.
- Set an `Allow` header for method-not-allowed responses.
- Keep GET and HEAD behavior compatible with `net/http`.

Non-goals:
- Do not implement runtime compression.
- Do not add ETag generation or router-level static policy.
- Do not change public option names.

Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
No docs change required unless response behavior descriptions change.

Done Definition:
- Compression negotiation follows token and quality semantics.
- Cache variance is explicit for precompressed-capable URLs.
- 405 responses advertise allowed methods.
- The listed validation commands pass.

Outcome:
- Added `Allow: GET, HEAD` to method-not-allowed responses.
- Updated precompressed negotiation to honor `Accept-Encoding` quality factors
  and wildcard handling.
- Added `Vary: Accept-Encoding` for uncompressed responses when a
  precompressed variant exists.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
