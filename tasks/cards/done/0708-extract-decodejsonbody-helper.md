# Card 0708

Priority: P2

Goal:
- Extract a shared internal JSON decode helper to eliminate the duplicated
  decode pipeline between `BindJSON` and `BindJSONWithOptions`.

Problem:

`BindJSON` (context_bind.go:23-52) and `BindJSONWithOptions` (context_bind.go:55-86)
share identical logic for four steps: empty-body check, decoder construction,
JSON decode, and trailing-data check. Only one line differs — whether
`DisallowUnknownFields` is set on the decoder.

```go
// In BindJSON (lines 33-50):
if len(bytes.TrimSpace(data)) == 0 {
    return &BindError{...ErrEmptyRequestBody...}
}
decoder := json.NewDecoder(bytes.NewReader(data))
if err := decoder.Decode(dst); err != nil {
    return &BindError{...ErrInvalidJSON...}
}
if decoder.Decode(&struct{}{}) != io.EOF {
    return &BindError{...ErrUnexpectedExtraData...}
}

// In BindJSONWithOptions (lines 68-84): identical except line 73-75:
decoder := json.NewDecoder(bytes.NewReader(data))
if opts.DisallowUnknownFields {
    decoder.DisallowUnknownFields()
}
```

Any fix to one path (error messages, sentinel wrapping, trailing-data handling)
must be manually applied to the other. This has already caused divergence: the
`joinSentinel` wrapper is used in both but the empty-body sentinel in
`BindJSONWithOptions` (line 69) uses a direct `ErrEmptyRequestBody` reference
while `BindJSON` (line 34) also does — currently consistent, but only by accident.

Fix: Extract an unexported helper:
```go
func decodeJSONBody(data []byte, dst any, disallowUnknown bool) error
```
Both `BindJSON` and `BindJSONWithOptions` call it; the only external difference
is the `disallowUnknown` flag.

Non-goals:
- Do not change `BindJSON` or `BindJSONWithOptions` signatures.
- Do not change observable error types or messages.
- Do not touch `BindAndValidateJSON` or the query variants.

Files:
- `contract/context_bind.go`

Tests:
- Existing tests must continue to pass unchanged.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- A single `decodeJSONBody` helper contains the empty-check, decode, and
  trailing-check logic.
- `BindJSON` and `BindJSONWithOptions` delegate to it.
- All existing tests pass.

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
