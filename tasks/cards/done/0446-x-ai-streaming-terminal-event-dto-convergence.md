# Card 0446: x/ai/streaming Terminal Event DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: x/ai/streaming
Owned Files:
- `x/ai/streaming/handler.go`
- `x/ai/streaming/handler_test.go`
- `docs/modules/x-ai/README.md`
Depends On: none

Goal:
Converge streaming handler terminal SSE payloads on local typed DTO structs and
test the final event shapes explicitly.

Problem:
`x/ai/streaming/handler.go` still builds the `complete` and `result` terminal
SSE events with ad hoc map literals before marshaling them to JSON. The
corresponding success tests only assert that some output exists, so accidental
field drift in these user-facing streaming events would be easy to miss.

Scope:
- Replace terminal SSE event maps with local typed DTO structs.
- Keep event names and JSON fields unchanged: `event`, `message`, `success`,
  and `results_count`.
- Add focused tests that parse the terminal SSE events and decode them into
  typed payloads.
- Update the x/ai primer with the streaming terminal DTO policy.

Non-goals:
- Do not change `WorkflowRequest`, workflow parsing, orchestration behavior, or
  SSE stream framing.
- Do not change `x/ai/sse`.
- Do not add dependencies.

Files:
- `x/ai/streaming/handler.go`
- `x/ai/streaming/handler_test.go`
- `docs/modules/x-ai/README.md`

Tests:
- `go test -race -timeout 60s ./x/ai/streaming/...`
- `go test -timeout 20s ./x/ai/streaming/...`
- `go vet ./x/ai/streaming/...`

Docs Sync:
Update `docs/modules/x-ai/README.md` to state that streaming terminal SSE
payloads use local typed DTO structs.

Done Definition:
- Terminal `complete` and `result` events are marshaled from typed structs.
- Success-path handler tests decode the terminal SSE payloads into typed
  structs.
- The listed validation commands pass.

Outcome:
- Added local typed terminal SSE DTOs for `complete` and `result` events.
- Replaced ad hoc terminal event map marshaling in the streaming handler.
- Added success-path tests that parse the SSE frame data and decode terminal
  payloads into typed structs.
- Documented the x/ai streaming terminal DTO policy.

Validation:
- `go test -race -timeout 60s ./x/ai/streaming/...`
- `go test -timeout 20s ./x/ai/streaming/...`
- `go vet ./x/ai/streaming/...`
