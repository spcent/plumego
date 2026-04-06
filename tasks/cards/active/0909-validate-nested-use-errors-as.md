# Card 0909

Priority: P3
State: active
Primary Module: contract
Owned Files:
- `contract/validation.go`
Depends On: —

Goal:
- Replace the bare type assertion `err.(validationErrors)` in `validateNestedStructField` with `errors.As` to make error extraction robust and idiomatic.

Problem:
`validateNestedStructField` calls `validateStructAtDepth` and extracts field errors via
a direct type assertion:

```go
// lines 117-120 and 123-128
if err := validateStructAtDepth(value.Interface(), fieldName, depth); err != nil {
    if issues, ok := err.(validationErrors); ok {
        return issues
    }
    // non-validationErrors silently discarded ↑
}
```

Two problems:

1. **Silent discard.** If `validateStructAtDepth` ever returns an error that is not
   exactly `validationErrors` (e.g., a wrapped version, a future alternative), it is
   silently dropped. Callers would see no validation errors when errors actually occurred.

2. **Not idiomatic.** The Go standard pattern for extracting a concrete error type from
   the error chain is `errors.As`, which traverses wrapping. The bare assertion skips any
   wrapping layer.

`validateStructAtDepth` currently always returns `validationErrors` or `nil`, so no bug
exists today — but the fragile assertion is a maintenance trap. Future changes to error
propagation (e.g., adding a wrapping layer for depth context) would silently break
validation without any compile-time or test signal.

Fix:
```go
var issues validationErrors
if err := validateStructAtDepth(...); errors.As(err, &issues) {
    return issues
}
```

Apply to both Ptr and Struct branches in `validateNestedStructField`.

Non-goals:
- Do not change `validationErrors` or `ValidateStruct`.
- Do not change validation rules or error messages.

Files:
- `contract/validation.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- Both branches in `validateNestedStructField` use `errors.As` instead of direct type assertion.
- No bare `.(validationErrors)` assertion exists in the file.
- All tests pass.

Outcome:
