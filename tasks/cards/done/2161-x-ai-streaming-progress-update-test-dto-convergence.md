# Card 2161: x/ai/streaming Progress Update Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: done
Primary Module: x/ai/streaming
Owned Files:
- `x/ai/streaming/streaming_test.go`
Depends On: none

Goal:
Make the progress update JSON shape test decode into a typed payload instead
of a generic map.

Problem:
`TestProgressUpdate/MarshalJSON` still decodes the marshaled progress update
into `map[string]any` and manually asserts the timestamp type. This is weaker
than the nearby typed DTO tests and does not clearly lock the expected JSON
field shape.

Scope:
- Replace the map decode with a local typed assertion struct.
- Keep the tested behavior unchanged.

Non-goals:
- Do not change `ProgressUpdate` or streaming runtime behavior.
- Do not add dependencies.

Files:
- `x/ai/streaming/streaming_test.go`

Tests:
- `go test -race -timeout 60s ./x/ai/streaming/...`
- `go test -timeout 20s ./x/ai/streaming/...`
- `go vet ./x/ai/streaming/...`

Docs Sync:
No docs change required; this is test-only contract tightening.

Done Definition:
- The progress update test no longer decodes through `map[string]any`.
- The listed validation commands pass.

Outcome:
- Replaced the progress update JSON test's generic map decode with a local
  typed assertion struct.
- Added field assertions for workflow, step, status, progress, and timestamp.

Validation:
- `go test -race -timeout 60s ./x/ai/streaming/...`
- `go test -timeout 20s ./x/ai/streaming/...`
- `go vet ./x/ai/streaming/...`
