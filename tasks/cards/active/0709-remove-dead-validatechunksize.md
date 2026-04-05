# Card 0709

Priority: P3

Goal:
- Remove the dead `validateChunkSize` function from `context_stream.go` and
  fix the semantic mismatch between it and the live chunk-size check in `Stream`.

Problem:

`context_stream.go:332-336`:
```go
func validateChunkSize(chunkSize int) error {
    if chunkSize <= 0 {
        return ErrInvalidChunkSize
    }
    return nil
}
```

This function is never called. The actual chunk-size guard is inlined in
`Stream` (context_stream.go:55):
```go
if cfg.ChunkSize < 0 {
    return ErrInvalidChunkSize
}
```

The two checks have different semantics:
- `validateChunkSize` rejects `chunkSize == 0` (treats zero as invalid).
- `Stream` allows `chunkSize == 0` (zero means "no chunking").

Zero is a valid and meaningful value for `ChunkSize` (it disables chunked
flushing). A function that would reject zero is wrong for this API. Because
`validateChunkSize` was never wired in, the bug never surfaced — but leaving
dead code with incorrect semantics is a trap for the next reader.

Scope:
- Delete `validateChunkSize` entirely.
- Verify no call site exists: `grep -rn 'validateChunkSize' . --include='*.go'`

Non-goals:
- Do not change the `Stream` chunk-size check.
- Do not add a replacement helper.

Files:
- `contract/context_stream.go`

Tests:
- `go test ./contract/...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `validateChunkSize` does not exist.
- `grep` for the name returns no results.
- All tests pass.
