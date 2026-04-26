# Card 2267

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: active
Primary Module: x/gateway
Owned Files:
- x/gateway/example_test.go
- docs/modules/x-gateway/README.md
Depends On: 2266

Goal:
Add a runnable edge-gateway example that shows explicit proxy registration and safe dynamic configuration.

Scope:
- Add an offline example using `httptest` to register a proxy route to a backend.
- Prefer error-returning construction paths for dynamic inputs.
- Update the gateway primer to point to the example and clarify discovery remains caller-owned.

Non-goals:
- Do not add discovery defaults.
- Do not add a new protocol adapter.
- Do not change gateway runtime behavior.

Files:
- `x/gateway/example_test.go`
- `docs/modules/x-gateway/README.md`

Tests:
- `go test -timeout 20s ./x/gateway/...`
- `go vet ./x/gateway/...`

Docs Sync:
- Required in the `x/gateway` primer.

Done Definition:
- The example runs offline through `go test`.
- It teaches explicit edge wiring without hiding backend or discovery ownership.

Outcome:
