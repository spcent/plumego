# Card 2224

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/cors/cors.go; middleware/cors/cors_test.go; tasks/cards/done/2224-middleware-cors-request-header-strictness.md
Depends On:

Goal:
Make CORS preflight reject disallowed requested headers instead of returning a partially acknowledged preflight response.

Scope:
- Require every requested preflight header to be allowed when `AllowedHeaders` is not wildcard.
- Keep case-insensitive matching and preserve original requested header casing in `Access-Control-Allow-Headers`.
- Pass through disallowed preflight header requests without CORS allow headers.
- Keep existing wildcard and no-request-header behavior.

Non-goals:
- Do not add route-aware CORS policy.
- Do not change origin or method matching semantics.
- Do not introduce a new constructor or dependency.

Files:
- `middleware/cors/cors.go`
- `middleware/cors/cors_test.go`

Tests:
- `go test -timeout 20s ./middleware/cors`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this tightens the existing preflight contract.

Done Definition:
- Disallowed requested headers do not receive CORS allow headers.
- Allowed requested headers still return 204 with canonical vary headers.
- Targeted middleware tests and vet pass.

Outcome:
- Added all-or-nothing requested-header validation for CORS preflight.
- Delayed writing allow headers until method and header validation both pass.
- Added regression tests for disallowed requested headers and wildcard header echoing.
- Validation run: `go test -timeout 20s ./middleware/cors`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
