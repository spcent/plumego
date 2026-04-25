# Card 2164: x/gateway/protocolmw Error Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: active
Primary Module: x/gateway/protocolmw
Owned Files:
- `x/gateway/protocolmw/middleware_test.go`
Depends On: none

Goal:
Converge protocol middleware error tests on typed error response decoding.

Problem:
Protocol middleware tests still decode structured error responses through
generic maps. The implementation already writes contract-shaped transport
errors with stable codes and stage details, so the tests should assert that
shape explicitly.

Scope:
- Replace generic error response map decodes with a local typed response struct.
- Keep existing error code, status, body-redaction, and stage assertions.

Non-goals:
- Do not change protocol middleware behavior or public APIs.
- Do not add dependencies.

Files:
- `x/gateway/protocolmw/middleware_test.go`

Tests:
- `go test -race -timeout 60s ./x/gateway/protocolmw/...`
- `go test -timeout 20s ./x/gateway/protocolmw/...`
- `go vet ./x/gateway/protocolmw/...`

Docs Sync:
No docs change required; this is test-only contract tightening.

Done Definition:
- Protocol middleware error helpers no longer decode through `map[string]any`.
- The listed validation commands pass.

Outcome:
