# Card 0729

Priority: P2

Goal:
- Prevent `ErrorLogger` from silently overwriting reserved log field keys
  (`code`, `status`, `category`, `trace_id`) when an `APIError.Details` map
  contains those same keys.

Problem:

`errors.go:223-239`:
```go
fields := log.Fields{
    "code":     err.Code,
    "status":   err.Status,
    "category": err.Category,
}
// ...
if err.TraceID != "" {
    fields["trace_id"] = err.TraceID
}

for k, v := range err.Details {
    fields[k] = v    // ← overwrites "code", "status", "category", "trace_id"
}
```

Any `APIError` whose `Details` map contains the keys `"code"`, `"status"`,
`"category"`, or `"trace_id"` will silently replace the structured log fields
with whatever is in `Details`. The structured values are computed from the
error's own typed fields (e.g. `err.Status int`, `err.Code string`), but they
can be silently replaced by arbitrary `Details` values of any type, breaking
log parsers that expect consistent field types.

Example:
```go
err := NewErrorBuilder().
    Status(400).
    Code("BAD_REQUEST").
    Detail("status", "custom-override").   // Detail wins over err.Status
    Build()
ErrorLogger(logger, r, err)
// log: status="custom-override" (was int 400)
```

Fix: Iterate `Details` first, then apply the typed fields so they always win:
```go
for k, v := range err.Details {
    fields[k] = v
}
fields["code"] = err.Code
fields["status"] = err.Status
fields["category"] = err.Category
if err.TraceID != "" {
    fields["trace_id"] = err.TraceID
}
```

Or alternatively, namespace `Details` under a sub-key:
```go
if len(err.Details) > 0 {
    fields["details"] = err.Details
}
```
This avoids the collision entirely and keeps structured fields predictable.
Namespacing under `"details"` is preferred because it mirrors the JSON envelope
structure (`{"error": {"details": {...}}}`) and is unambiguous in log queries.

Non-goals:
- Do not change what keys `ErrorBuilder` accepts in `Details`.
- Do not change the JSON serialization of `APIError`.

Files:
- `contract/errors.go`

Tests:
- Add a test: `ErrorLogger` with `Details["code"] = "OVERRIDE"` must log
  `code = err.Code`, not `"OVERRIDE"`.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `ErrorLogger` structured fields (`code`, `status`, `category`, `trace_id`)
  are never overwritten by `Details` values.
- All tests pass.
