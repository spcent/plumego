# Card 0739

Priority: P3
State: active
Primary Module: contract
Owned Files: contract/context_stream.go

Goal:
- Consolidate the two standalone sentinel `var` declarations in
  `context_stream.go` into `var(...)` blocks so their style matches every
  other sentinel error in the package.

Problem:

`context_stream.go:16`:
```go
var ErrUnsupportedStreamSource = errors.New("unsupported StreamSource type")
```

`context_stream.go:280`:
```go
var ErrSSENotSupported = errors.New("SSE not supported: response writer does not implement http.Flusher")
```

Both are lone `var` statements floating at the top of their declaration sites.
Every other sentinel in the package (`context_core.go:80-125`,
`context_response.go:12-15`, `auth.go`) uses a `var(...)` block, even when
only a single error is declared. The inconsistency creates unnecessary visual
noise and makes it harder to spot new errors added to the file.

Fix for `ErrUnsupportedStreamSource` (near top of file, after imports):
```go
var (
    // ErrUnsupportedStreamSource is returned when Stream receives a Source value
    // whose type does not match any of the accepted streaming primitives.
    ErrUnsupportedStreamSource = errors.New("unsupported StreamSource type")
)
```

Fix for `ErrSSENotSupported` (near `SSEWriter` type, line ~280):
```go
var (
    // ErrSSENotSupported is returned when the response writer doesn't support SSE.
    ErrSSENotSupported = errors.New("SSE not supported: response writer does not implement http.Flusher")
)
```

Non-goals:
- Do not rename, move, or reorder sentinel errors between files.
- Do not change any logic or error messages.

Files:
- `contract/context_stream.go`

Tests:
- No behaviour change; existing tests pass unchanged.
- `go build ./...`
- `go vet ./contract/...`

Done Definition:
- Both `ErrUnsupportedStreamSource` and `ErrSSENotSupported` are declared
  inside `var(...)` blocks.
- No standalone `var` sentinel declarations remain in `context_stream.go`.
- All tests pass.

Outcome:
- Completed by consolidating both sentinel declarations in
  `contract/context_stream.go` into `var (...)` blocks without changing error
  messages or behavior.

Validation Run:
- `gofmt -w contract/context_stream.go`
- `rg -n "var ErrUnsupportedStreamSource|var ErrSSENotSupported" contract/context_stream.go`
- `go vet ./contract/...`
- `go build ./...`
