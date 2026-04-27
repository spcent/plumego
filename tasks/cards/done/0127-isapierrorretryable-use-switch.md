# Card 0127

Priority: P3
State: done
Primary Module: contract
Owned Files: contract/errors.go

Goal:
- Replace the per-call slice allocation in `IsAPIErrorRetryable` with a
  `switch` statement.

Problem:

`errors.go:459-471`:
```go
func IsAPIErrorRetryable(err APIError) bool {
    retryableCodes := []int{408, 429, 500, 502, 503, 504}   // ← heap-allocated on every call

    for _, code := range retryableCodes {
        if err.Status == code {
            return true
        }
    }

    return err.Category == CategoryTimeout
}
```

`IsAPIErrorRetryable` is called by `IsRetryable` (`error_wrap.go:116`), which
is called on every error in middleware retry loops. Allocating a six-element
slice on each invocation is unnecessary and violates the zero-allocation
principle for hot-path helpers. A `switch` statement is both zero-allocation
and idiomatic Go for a small, static set of values.

Fix:
```go
func IsAPIErrorRetryable(err APIError) bool {
    switch err.Status {
    case 408, 429, 500, 502, 503, 504:
        return true
    }
    return err.Category == CategoryTimeout
}
```

Non-goals:
- Do not change which status codes are considered retryable.
- Do not change `IsRetryable` in `error_wrap.go`.
- Do not add new retryable conditions.

Files:
- `contract/errors.go`

Tests:
- Existing tests cover the retryable codes. Run them unchanged.
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `IsAPIErrorRetryable` uses a `switch` with no slice allocation.
- All existing retryable-code tests pass.

Outcome:
- Completed by replacing the per-call retryable-status slice with a `switch`
  over the same fixed set of HTTP status codes.

Validation Run:
- `gofmt -w contract/errors.go`
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
