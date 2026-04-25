# Card 2167: x/resilience/circuitbreaker Error Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: active
Primary Module: x/resilience/circuitbreaker
Owned Files:
- `x/resilience/circuitbreaker/circuitbreaker_test.go`
Depends On: none

Goal:
Make the circuit breaker middleware open-circuit response test decode a typed
error response shape.

Problem:
The middleware test for open circuit responses decodes a fixed structured error
payload through nested `map[string]any` assertions. This hides the response
shape that the test is trying to lock.

Scope:
- Replace nested map decoding with a local typed error response struct.
- Keep existing status, code, circuit detail, and header assertions.

Non-goals:
- Do not change circuit breaker behavior or public APIs.
- Do not add dependencies.

Files:
- `x/resilience/circuitbreaker/circuitbreaker_test.go`

Tests:
- `go test -race -timeout 60s ./x/resilience/circuitbreaker/...`
- `go test -timeout 20s ./x/resilience/circuitbreaker/...`
- `go vet ./x/resilience/circuitbreaker/...`

Docs Sync:
No docs change required; this is test-only contract tightening.

Done Definition:
- The open-circuit middleware response test no longer decodes through nested
  maps.
- The listed validation commands pass.

Outcome:
