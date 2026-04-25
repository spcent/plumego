# Card 2169: Middleware Conformance Error DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
- `middleware/conformance_test.go`
- `middleware/conformance/helpers_test.go`
Depends On: none

Goal:
Make middleware conformance error-envelope assertions decode a typed response
shape instead of nested generic maps.

Problem:
The middleware conformance tests validate a fixed canonical error envelope, but
the helpers decode the response through `map[string]any` and nested type
assertions. That keeps required fields implicit and differs from the typed
response assertions used by nearby modules.

Scope:
- Add local typed error-envelope DTOs for the conformance assertion helpers.
- Preserve existing status, content-type, required-field, and code assertions.

Non-goals:
- Do not change middleware behavior or public APIs.
- Do not change tenant/auth/rate-limit/recovery code paths.
- Do not add dependencies.

Files:
- `middleware/conformance_test.go`
- `middleware/conformance/helpers_test.go`

Tests:
- `go test -race -timeout 60s ./middleware/...`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
No docs change required; this is a test-only conformance cleanup.

Done Definition:
- Middleware conformance error helpers no longer decode fixed envelopes through
  nested maps.
- The listed validation commands pass.

Outcome:
