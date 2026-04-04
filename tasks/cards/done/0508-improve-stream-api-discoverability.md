# Card 0508

Priority: P1

Goal:
- Replace the 15-method stream API explosion on `Ctx` with a composable
  `Stream(cfg StreamConfig) error` entry point, reducing API surface while
  preserving all existing streaming modes.

Problem:
- `context_stream.go` exposes 15 methods following the pattern
  `Stream{JSON|Text|SSE}[WithChannel|WithGenerator|Chunked|WithRetry]()`.
- The naming is hard to discover; combinations like "Channel + Retry" are not
  possible despite being useful.
- Each method is a full independent implementation, leading to code duplication
  for header-setting, flushing, and error propagation.

Proposed shape:
```go
type StreamConfig struct {
    Format    StreamFormat       // JSON | Text | SSE
    Source    StreamSource       // slice, channel, generator
    ChunkSize int                // optional, 0 = no chunking
    MaxRetry  int                // optional, 0 = no retry
}

func (c *Ctx) Stream(cfg StreamConfig) error
```

Scope:
- Define `StreamConfig`, `StreamFormat`, `StreamSource` types.
- Implement `(c *Ctx).Stream(cfg StreamConfig) error` routing to shared
  internal helpers.
- Keep existing methods as thin wrappers calling `Stream()` internally so
  callers are not broken, but mark them as deprecated via a comment.
- Extract shared header/flush logic into one unexported helper.

Non-goals:
- Do not remove the existing 15 methods in this card — leave them as
  deprecated wrappers. Removal is a follow-up card after callers migrate.
- Do not change the wire format of SSE, NDJSON, or plain-text streams.

Files:
- `contract/context_stream.go`

Tests:
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `Stream(cfg StreamConfig) error` exists and supports all three formats and
  all four source modes (slice, channel, generator, chunked).
- Existing `Stream*` methods delegate to `Stream()`.
- All existing stream tests pass against the new implementation.
