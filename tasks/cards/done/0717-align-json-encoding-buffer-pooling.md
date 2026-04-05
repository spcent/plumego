# Card 0717

Priority: P3

Goal:
- Make JSON buffer pooling consistent: `WriteJSON` should use the same pooled
  buffer that `Ctx.JSON` already uses, so the canonical response helpers don't
  allocate more than the low-level helper.

Problem:

`Ctx.JSON` (context_response.go:42-57) uses a pooled `bytes.Buffer`:
```go
func (c *Ctx) JSON(status int, data any) error {
    c.W.Header().Set("Content-Type", "application/json")
    c.W.WriteHeader(status)

    buf := getJSONBuffer()        // ← pooled
    defer putJSONBuffer(buf)

    if err := json.NewEncoder(buf).Encode(data); err != nil {
        return err
    }
    _, err := c.W.Write(buf.Bytes())
    return err
}
```

`WriteJSON` (response.go:17-24) encodes directly to the writer, allocating a
new encoder on every call:
```go
func WriteJSON(w http.ResponseWriter, status int, payload any) error {
    // ...
    return json.NewEncoder(w).Encode(payload)   // ← no pooling
}
```

`WriteResponse` delegates to `WriteJSON`, so `Ctx.Response` (which callers are
told to prefer in new code) allocates more than `Ctx.JSON`. The optimization is
applied to the lower-level helper but not to the canonical one.

Additionally, encoding to a pooled buffer before writing has a correctness
advantage: if encoding fails, no partial output has been written to the response
writer. The current `WriteJSON` would write partial JSON on encode error.

Fix: Have `WriteJSON` encode to a pooled buffer and then write:
```go
func WriteJSON(w http.ResponseWriter, status int, payload any) error {
    if w == nil {
        return ErrResponseWriterNil
    }
    buf := getJSONBuffer()
    defer putJSONBuffer(buf)
    if err := json.NewEncoder(buf).Encode(payload); err != nil {
        return err
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _, err := w.Write(buf.Bytes())
    return err
}
```

Note: this also fixes the current bug where headers are set before encoding,
meaning a failed encode has already written the status line.

Non-goals:
- Do not change `Ctx.JSON` (it already pools correctly).
- Do not change `WriteResponse` or `Ctx.Response` (they inherit the fix).
- Do not add a separate buffer pool; reuse the existing `jsonBufferPool`.

Files:
- `contract/response.go`
- `contract/response_buffer.go` (pool lives here; no change needed)

Tests:
- Add a test: encoding an unencodable value must not write partial output to
  the response writer.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `WriteJSON` uses `getJSONBuffer`/`putJSONBuffer`.
- `WriteJSON` sets headers only after successful encode.
- Existing tests pass; partial-write test added.

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
