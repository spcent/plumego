# Card 2265

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: active
Primary Module: x/rest
Owned Files:
- x/rest/example_test.go
- docs/modules/x-rest/README.md
Depends On: 2264

Goal:
Add a compact runnable `x/rest` example that shows explicit resource route registration without implying app bootstrap ownership.

Scope:
- Add an offline example test for `ResourceSpec`, repository-backed controller construction, and explicit router registration.
- Use `contract` response conventions already owned by `x/rest`.
- Update the primer to point users to the runnable example.

Non-goals:
- Do not add a new `reference/with-rest` app in this card.
- Do not change `x/rest` public APIs.
- Do not add database dependencies.

Files:
- `x/rest/example_test.go`
- `docs/modules/x-rest/README.md`

Tests:
- `go test -timeout 20s ./x/rest/...`
- `go vet ./x/rest/...`

Docs Sync:
- Required in the `x/rest` primer.

Done Definition:
- `go test ./x/rest/...` runs the example.
- The example demonstrates route binding as app-local explicit wiring.

Outcome:
