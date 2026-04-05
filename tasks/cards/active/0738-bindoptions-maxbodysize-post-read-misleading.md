# Card 0738

Priority: P3

Goal:
- Clarify in `BindOptions` and `BindJSONWithOptions` that `MaxBodySize` is a
  post-read cap, not a read-time limit, so callers understand when to use it
  versus `RequestConfig.MaxBodySize`.

Problem:

`bind_helpers.go:8-16`:
```go
type BindOptions struct {
    // MaxBodySize caps the body size checked after the initial read.
    MaxBodySize int64
    // ...
}
```

`context_bind.go:64-66`:
```go
if opts.MaxBodySize > 0 && int64(len(data)) > opts.MaxBodySize {
    return &BindError{...ErrRequestBodyTooLarge...}
}
```

The check happens after `c.bodyBytes()` has already read the complete body
into memory. This means:

1. **Memory is NOT protected**: If `RequestConfig.MaxBodySize` is 10 MB and
   `opts.MaxBodySize` is 1 KB, the full 10 MB body is read into memory first.
   The 1 KB limit is then applied and the data is discarded. The memory
   allocation already happened.

2. **Name implies read-time enforcement**: `MaxBodySize` reads as "maximum
   body size" which conventionally means a read-time limit. Callers who set
   it to protect against large payloads will be surprised to learn the data
   was already allocated.

3. **Redundancy with `RequestConfig.MaxBodySize`**: `RequestConfig.MaxBodySize`
   IS a read-time limit (enforced via `http.MaxBytesReader` in `bodyBytes()`).
   `opts.MaxBodySize` only adds value when you want a stricter per-endpoint
   limit without changing the global config — but this use case is not
   documented.

Fix: Update the doc comments to be explicit:

```go
type BindOptions struct {
    // MaxBodySize is a per-call size limit applied after the body has been
    // read into memory. It does not prevent memory allocation for bodies
    // within the global RequestConfig.MaxBodySize limit.
    // Use RequestConfig.MaxBodySize for read-time protection; use this field
    // only when a stricter per-endpoint limit is needed post-read.
    MaxBodySize int64
```

And in `BindJSONWithOptions`:
```go
// opts.MaxBodySize, if positive, enforces a stricter per-call cap on the
// already-read body. The global RequestConfig.MaxBodySize enforces read-time limits.
```

Non-goals:
- Do not change the behavior of the check.
- Do not move the check to read time (that would require duplicating
  `bodyBytes` logic and breaking body caching).
- Do not remove `MaxBodySize` from `BindOptions`.

Files:
- `contract/bind_helpers.go`
- `contract/context_bind.go`

Tests:
- No behavior change; existing tests pass unchanged.
- `go build ./...`

Done Definition:
- `BindOptions.MaxBodySize` and the `BindJSONWithOptions` implementation both
  have comments that accurately describe the post-read semantics.
- All tests pass.
