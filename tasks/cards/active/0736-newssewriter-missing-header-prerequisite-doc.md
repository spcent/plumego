# Card 0736

Priority: P3

Goal:
- Document that `NewSSEWriter` requires SSE response headers to have been set
  before it is called, or enforce the requirement in code.

Problem:

`context_stream.go:282-290`:
```go
// NewSSEWriter creates a new SSE writer.
// Returns an error if the response writer doesn't support flushing.
func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
    f, ok := w.(http.Flusher)
    if !ok {
        return nil, ErrSSENotSupported
    }
    return &SSEWriter{w: w, f: f}, nil
}
```

`NewSSEWriter` is exported and creates an `SSEWriter` from any
`http.ResponseWriter`. However, it does not set the three mandatory SSE
response headers:
- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`

The correct usage sequence is:
1. Call `SetSSEHeaders()` (or `RespondWithSSE()`) to write headers.
2. Then call `NewSSEWriter()` to get a writer.

But `NewSSEWriter`'s doc comment says nothing about this prerequisite. A
caller who calls `NewSSEWriter` directly (e.g., when building a custom SSE
response flow) will create an `SSEWriter` that writes events without the
required headers, causing clients to misinterpret the stream.

`RespondWithSSE()` is the safe entry point that does both steps atomically, but
`NewSSEWriter` is exported without guidance on when it should be used.

Fix:

**Option A (preferred): Update doc comment**
```go
// NewSSEWriter creates an SSEWriter for the given response writer.
// The caller is responsible for setting SSE headers before writing events;
// use Ctx.RespondWithSSE() which sets headers and creates the writer atomically.
// Returns ErrSSENotSupported if the response writer does not support flushing.
func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error)
```

**Option B: Make NewSSEWriter unexported**
If `NewSSEWriter` has no legitimate callers outside the package, unexport it
to `newSSEWriter`. Callers should use `Ctx.RespondWithSSE()` exclusively.

Check callers: `grep -rn 'NewSSEWriter' . --include='*.go'`

If no external callers exist, Option B is preferred. Otherwise, Option A.

Non-goals:
- Do not change `RespondWithSSE` behavior.
- Do not merge header-setting into `NewSSEWriter` (that changes the API
  contract and affects `WriteSSE`).

Files:
- `contract/context_stream.go`

Tests:
- `go build ./...`
- `go vet ./...`

Done Definition:
- `NewSSEWriter`'s doc comment (Option A) or visibility (Option B) makes the
  header-setup prerequisite unambiguous.
- All tests pass.
