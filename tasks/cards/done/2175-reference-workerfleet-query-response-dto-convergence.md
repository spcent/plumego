# Card 2175: Reference Workerfleet Query Response DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: reference/workerfleet
Owned Files:
- `reference/workerfleet/internal/handler/query_test.go`
Depends On: none

Goal:
Make workerfleet query handler tests decode fixed response data into typed DTOs.

Problem:
Two workerfleet handler tests unmarshal the response into `contract.Response`,
then assert `Data` through `map[string]any`. The handler responses have fixed
application DTO shapes, so the map assertions hide the intended contract and
require float casts for numeric fields.

Scope:
- Add a local generic response envelope test helper/shape.
- Decode heartbeat and timeline success responses into typed data structs.
- Preserve existing status and field assertions.

Non-goals:
- Do not change workerfleet handler behavior or public APIs.
- Do not change `contract.Response`.
- Do not add dependencies.

Files:
- `reference/workerfleet/internal/handler/query_test.go`

Tests:
- from `reference/workerfleet`: `go test -race -timeout 60s ./internal/handler/...`
- from `reference/workerfleet`: `go test -timeout 20s ./internal/handler/...`
- from `reference/workerfleet`: `go vet ./internal/handler/...`

Docs Sync:
No docs change required; this is test-only response DTO convergence.

Done Definition:
- The targeted workerfleet success response tests no longer assert `Data`
  through `map[string]any`.
- The listed validation commands pass.

Outcome:
