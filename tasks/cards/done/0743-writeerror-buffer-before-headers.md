# Card 0743

Priority: P2
State: done
Primary Module: contract
Owned Files: contract/errors.go

Goal:
- Buffer the JSON encoding in `WriteError` so that an encoding failure does
  not leave the HTTP connection in an indeterminate state with headers already
  committed.

Problem:

`errors.go:192-198`:
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(err.Status)                            // ← headers committed here

return json.NewEncoder(w).Encode(ErrorResponse{     // ← if this fails...
    Error:   err,
    TraceID: err.TraceID,
})
```

Once `w.WriteHeader(err.Status)` is called, the HTTP status and headers are
sent to the client. If `json.NewEncoder(w).Encode(...)` then fails (e.g.,
`err.Details` contains an unexportable or unencodable value), the client
receives a response with the correct status code but an **empty or truncated
body**. The caller gets back the encoding error, but the damage is already done.

Compare with `WriteJSON` (`response.go:17-33`) and `Ctx.JSON`
(`context_response.go:42-57`), which both:
1. Encode the payload into a pooled `bytes.Buffer` first.
2. If encoding fails, return the error **before** writing any headers.
3. Only then set Content-Type, call `w.WriteHeader`, and write the buffer.

`WriteError` is the only JSON-writing function in the package that commits
headers before encoding. This inconsistency means any `Details` map entry
holding an unencodable type (channel, function pointer, etc.) produces a
silent partial response in production.

Fix — restructure `WriteError` to match `WriteJSON`:
```go
func WriteError(w http.ResponseWriter, r *http.Request, err APIError) error {
    // ... existing validation / fallback fills ...

    resp := ErrorResponse{Error: err, TraceID: err.TraceID}

    buf := getJSONBuffer()
    defer putJSONBuffer(buf)
    if encErr := json.NewEncoder(buf).Encode(resp); encErr != nil {
        return encErr   // nothing committed yet; caller decides how to proceed
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(err.Status)
    _, writeErr := w.Write(buf.Bytes())
    return writeErr
}
```

Non-goals:
- Do not change the `ErrorResponse` shape or `WriteError` signature.
- Do not change `ValidateError` or the fallback fill logic at the top of
  `WriteError`.
- Do not remove `WarnFunc` calls.

Files:
- `contract/errors.go`
- `contract/response_buffer.go` (no changes; `getJSONBuffer` / `putJSONBuffer`
  already exist there)

Tests:
- Add a test that passes an `APIError` whose `Details` map contains an
  unencodable value (e.g., a `chan int`) and verifies that `WriteError`
  returns a non-nil error **without** writing any bytes to the response
  writer:
  ```go
  rr := httptest.NewRecorder()
  err := WriteError(rr, nil, NewErrorBuilder().
      Status(400).Code("X").Message("m").Category(CategoryClient).
      Detail("bad", make(chan int)).     // unencodable
      Build())
  if err == nil {
      t.Fatal("expected encoding error, got nil")
  }
  if rr.Code != 200 {
      t.Fatalf("headers must not be written on encoding failure, got status %d", rr.Code)
  }
  ```
- Run existing error-writing tests to confirm no regressions.
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `WriteError` encodes into a buffer before committing headers.
- An encoding failure returns an error and writes nothing to `w`.
- All existing tests pass.

Outcome:
- Completed by buffering `WriteError` JSON output before setting headers and by
  adding a regression test that verifies encoding failures leave the recorder
  untouched.

Validation Run:
- `gofmt -w contract/errors.go contract/errors_test.go`
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
