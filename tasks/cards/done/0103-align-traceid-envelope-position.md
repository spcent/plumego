# Card 0103

Priority: P2

Goal:
- Make `trace_id` appear at the same nesting level in both success and error
  responses so clients can use a single JSON path to extract it.

Problem:

Success responses (via `WriteResponse` / `Ctx.Response`):
```json
{
  "data": { ... },
  "meta": { ... },
  "trace_id": "abc123"          ← top-level
}
```

Error responses (via `WriteError` / `Ctx.ErrorJSON`):
```json
{
  "error": {
    "code": "...",
    "message": "...",
    "category": "...",
    "trace_id": "abc123"        ← nested inside "error"
  }
}
```

A client that wants to log or display `trace_id` for every response must branch:
- Success: `body.trace_id`
- Error: `body.error.trace_id`

This adds boilerplate to every API client and breaks any generic middleware that
extracts `trace_id` from responses for logging.

Root cause: `Response` (response.go:10) carries `TraceID` as a top-level field,
while `APIError` (errors.go:141) carries it as an `APIError` field, and
`ErrorResponse` (errors.go:151) wraps `APIError` without promoting `TraceID`.

Two valid fixes:

**Option A (preferred): Promote trace_id to ErrorResponse**
```go
type ErrorResponse struct {
    Error   APIError `json:"error"`
    TraceID string   `json:"trace_id,omitempty"`
}
```
Move `APIError.TraceID` population to the `ErrorResponse` level in `WriteError`.
`APIError.TraceID` can remain for internal use but is omitted from JSON (`json:"-"`).

**Option B: Document as intentional**
Add a comment to both `Response` and `ErrorResponse` explaining the asymmetry
and why it is deliberate. This does not fix the client burden but at least
removes ambiguity for maintainers.

Option A is preferred: it aligns the wire format, reduces client complexity,
and is consistent with the `Response` struct's design.

Scope (Option A):
- Add `TraceID string \`json:"trace_id,omitempty"\`` to `ErrorResponse`.
- In `WriteError`, populate `ErrorResponse.TraceID` from `err.TraceID` (or
  context) and change `APIError.TraceID` json tag to `json:"-"`.
- Update `ParseErrorFromResponse` to read from `ErrorResponse.TraceID`.
- Update tests that assert on the error JSON shape.

Non-goals:
- Do not change the `data`/`meta` structure of success responses.
- Do not add trace_id to non-JSON error paths.

Files:
- `contract/errors.go`
- `contract/response.go` (for reference; no change expected)
- Tests: `contract/errors_test.go`, `contract/error_handling_test.go`

Tests:
- Add a test: `WriteError` output must have `trace_id` at the top of the JSON
  object, not inside `error`.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- Error responses have `trace_id` at the same level as the `error` key.
- Success responses continue to have `trace_id` at top level (no change).
- `ParseErrorFromResponse` reads trace_id from the correct location.
- All tests pass.

Outcome:
- Completed in the 2026-04-05 contract cleanup batch.
- Verified as part of the shared contract/task-card completion pass.

Validation Run:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
- `go build ./...`
- `go test -timeout 20s ./...`
- `go test -race -timeout 60s ./...`
- `go vet ./...`
