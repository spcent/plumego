# Card 0973: x/frontend Response Header Policy

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/config.go`
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0737

Goal:
Prevent caller-provided custom headers from overriding transport-critical file
serving semantics.

Scope:
- Reject or ignore unsafe custom header names such as hop-by-hop headers,
  `Content-Length`, `Content-Encoding`, `Transfer-Encoding`, and `Vary`.
- Keep security and metadata headers available through `WithHeaders`.
- Ensure internally computed `Content-Type`, `Content-Encoding`, `Vary`, and
  cache headers have deterministic precedence.
- Add regression coverage for disallowed and allowed headers.

Non-goals:
- Do not introduce a full CSP builder or security-header profile.
- Do not mutate caller-owned header maps.
- Do not change `contract.WriteError` behavior.

Files:
- `x/frontend/config.go`
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Document the disallowed custom response header list and precedence rules.

Done Definition:
- Unsafe custom headers cannot corrupt response framing or variant semantics.
- Allowed security headers still apply to file responses.
- Precompressed and custom MIME responses keep internal headers authoritative.
- The listed validation commands pass.

Outcome:
- `WithHeaders` now rejects transport-critical, cache, conditional, range,
  content, variant, and hop-by-hop headers during mount construction.
- Custom security and metadata headers still apply and caller-owned header maps
  remain isolated.
- Documentation now points callers to dedicated options for cache, MIME, and
  precompressed semantics.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
