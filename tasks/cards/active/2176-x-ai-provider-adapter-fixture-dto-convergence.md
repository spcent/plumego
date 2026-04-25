# Card 2176: x/ai Provider Adapter Fixture DTO Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: x/ai/provider
Owned Files:
- `x/ai/provider/adapters_test.go`
Depends On: none

Goal:
Make provider adapter mock API responses use typed fixtures instead of nested
generic maps.

Problem:
The Claude and OpenAI adapter tests build fixed mock provider responses with
deep `map[string]any` literals. Those responses have stable provider-specific
shapes, so the map literals are noisy and less readable than typed test
fixtures.

Scope:
- Add local typed mock response structs for Claude and OpenAI completion
  responses.
- Replace the fixed nested map fixtures in the successful completion tests.

Non-goals:
- Do not change provider adapter behavior or public APIs.
- Do not change request parsing, streaming, or error handling.
- Do not add dependencies.

Files:
- `x/ai/provider/adapters_test.go`

Tests:
- `go test -race -timeout 60s ./x/ai/provider/...`
- `go test -timeout 20s ./x/ai/provider/...`
- `go vet ./x/ai/provider/...`

Docs Sync:
No docs change required; this is test fixture cleanup.

Done Definition:
- Successful adapter tests no longer build fixed provider responses with nested
  generic maps.
- The listed validation commands pass.

Outcome:
