# Card 2172: x/gateway Transform Error Response DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: x/gateway
Owned Files:
- `x/gateway/transform/transform_test.go`
Depends On: none

Goal:
Make the gateway transform request-failure test decode its fixed error
response shape into a typed DTO.

Problem:
`TestRequestTransformFailureUsesGatewayTransformCode` still parses the
canonical error envelope through `map[string]any` and nested assertions. This is
an inconsistency with the typed gateway protocol middleware and transform tests
already introduced nearby.

Scope:
- Add or reuse a local typed error envelope DTO in the transform test file.
- Replace the remaining nested map decode in the request-transform failure
  assertion.

Non-goals:
- Do not change transform middleware behavior or public APIs.
- Do not change transformer callback signatures that intentionally accept JSON
  object maps.
- Do not add dependencies.

Files:
- `x/gateway/transform/transform_test.go`

Tests:
- `go test -race -timeout 60s ./x/gateway/transform/...`
- `go test -timeout 20s ./x/gateway/transform/...`
- `go vet ./x/gateway/transform/...`

Docs Sync:
No docs change required; this is test-only response DTO convergence.

Done Definition:
- The request-transform failure test no longer decodes its fixed envelope
  through nested maps.
- The listed validation commands pass.

Outcome:
