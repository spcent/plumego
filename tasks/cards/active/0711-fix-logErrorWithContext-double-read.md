# Card 0711

Priority: P3

Goal:
- Remove the redundant second param-extraction loop from `logErrorWithContext`
  in `error_utils.go`.

Problem:

`error_utils.go:84-114` — `logErrorWithContext`:

```go
func logErrorWithContext(logger log.StructuredLogger, r *http.Request, err error) {
    details := GetErrorDetails(err)           // ← extracts params here (line 89)
    fields := log.Fields{
        "operation": details["operation"],
        "module":    details["module"],
    }

    if params, ok := details["params"].(map[string]any); ok {
        for k, v := range params {            // ← spreads params into fields (line 96)
            fields[k] = v
        }
    }

    if wrappedErr, ok := err.(*WrappedErrorWithContext); ok {
        for k, v := range wrappedErr.Context.Params {  // ← reads same params AGAIN (line 102)
            fields[k] = v
        }
    }
    // ...
}
```

`GetErrorDetails` (error_wrap.go:141-183) already walks the `*WrappedErrorWithContext`
chain and puts `Context.Params` into `details["params"]` (line 156-158). The
block at lines 101-105 then reads `wrappedErr.Context.Params` directly and
iterates it again — setting the exact same keys in `fields`.

The second loop is a no-op when `GetErrorDetails` succeeds. It would only add
something if `GetErrorDetails` returned a different set of params, which cannot
happen for the same error value. The redundancy makes the code harder to follow
and misleads future readers into thinking there is a meaningful distinction.

Scope:
- Remove lines 101-105 (the `if wrappedErr, ok := err.(*WrappedErrorWithContext)` block).
- Verify that no params are lost by inspecting what `GetErrorDetails` already
  produces for `WrappedErrorWithContext`.

Non-goals:
- Do not change `GetErrorDetails`.
- Do not change logging output for any error type.

Files:
- `contract/error_utils.go`

Tests:
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- Lines 101-105 removed.
- `logErrorWithContext` reads params exactly once (via `GetErrorDetails`).
- Logging output for `*WrappedErrorWithContext` is unchanged.
- All tests pass.
