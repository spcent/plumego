# Card 0731: x/frontend Negotiation and Error Status Semantics

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
Depends On: 0730

Goal:
Make HTTP compression negotiation and custom error-page status handling stable
under cache, conditional, and range requests.

Scope:
- Parse `Accept-Encoding` once and choose supported precompressed variants by
  quality factor with deterministic server preference ties.
- Keep `br` preferred only when quality is equal to gzip.
- Prevent custom 404/5xx pages from returning 304 or 206 status codes.
- Add regression coverage for q-value ordering, invalid q values, conditional
  requests, and range requests on custom error pages.

Non-goals:
- Do not add ETag generation.
- Do not implement runtime compression.
- Do not change success asset conditional request behavior.

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
- `gzip;q=1, br;q=0.1` serves gzip when both variants exist.
- Equal q-values keep the existing br-over-gzip server preference.
- Custom error pages preserve their configured error status under conditional
  and range request headers.
- The listed validation commands pass.

Outcome:
- Precompressed negotiation now parses `Accept-Encoding` once and orders br/gzip
  by q-value, with br preferred only on equal quality.
- Invalid q-value tokens are ignored instead of being treated as accepted.
- Custom 404/5xx page serving strips conditional and range request headers so
  error pages retain their configured error status.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
