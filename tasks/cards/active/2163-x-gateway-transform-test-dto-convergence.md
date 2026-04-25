# Card 2163: x/gateway/transform Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: active
Primary Module: x/gateway/transform
Owned Files:
- `x/gateway/transform/transform_test.go`
Depends On: none

Goal:
Converge gateway transform tests on typed JSON assertion payloads where the
expected field shape is static.

Problem:
Several transform tests still encode or decode fixed JSON shapes through
`map[string]any`. That is appropriate for transformer callbacks, but less
precise for test fixtures and assertions where the expected payload is known.

Scope:
- Replace fixed response fixture maps with typed structs.
- Replace fixed response assertion maps with typed structs.
- Keep transformer callback maps unchanged because those are the package API.

Non-goals:
- Do not change transform behavior or public APIs.
- Do not add dependencies.

Files:
- `x/gateway/transform/transform_test.go`

Tests:
- `go test -race -timeout 60s ./x/gateway/transform/...`
- `go test -timeout 20s ./x/gateway/transform/...`
- `go vet ./x/gateway/transform/...`

Docs Sync:
No docs change required; this is test-only contract tightening.

Done Definition:
- Static transform test fixtures/assertions no longer use generic maps.
- The listed validation commands pass.

Outcome:
