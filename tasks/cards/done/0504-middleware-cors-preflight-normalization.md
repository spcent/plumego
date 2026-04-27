# Card 0504

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/cors/cors.go; middleware/cors/cors_test.go; tasks/cards/done/0504-middleware-cors-preflight-normalization.md
Depends On:

Goal:
Normalize CORS preflight behavior so the transport middleware only confirms explicitly allowed methods and headers and emits complete cache variation headers.

Scope:
- Apply CORS defaults through one helper instead of inline mutation in the constructor.
- Reject unmatched preflight requested methods by passing through instead of advertising methods that do not authorize the preflight.
- Add `Vary` coverage for `Origin`, `Access-Control-Request-Method`, and `Access-Control-Request-Headers`.
- Preserve existing successful CORS behavior for allowed origins, wildcard origins, credentials, and response bodies.

Non-goals:
- Do not add domain policy, tenant policy, or route-aware CORS behavior.
- Do not introduce a new dependency or a second CORS constructor family.
- Do not change non-CORS middleware.

Files:
- `middleware/cors/cors.go`
- `middleware/cors/cors_test.go`

Tests:
- `go test -timeout 20s ./middleware/cors`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; the change tightens existing CORS semantics.

Done Definition:
- Preflight tests prove disallowed methods are not acknowledged.
- Vary header tests cover preflight request method and headers.
- Targeted middleware tests and vet pass.

Outcome:
- Added a CORS defaulting helper and case-insensitive method matching for preflight checks.
- Preflight now adds `Vary` for origin, request method, and request headers.
- Disallowed preflight methods pass through without CORS allow headers.
- Validation run: `go test -timeout 20s ./middleware/cors`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
