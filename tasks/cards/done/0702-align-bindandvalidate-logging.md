# Card 0702

Priority: P2

Goal:
- Make error logging consistent across all four `BindAndValidate*` methods.
  Currently only `BindAndValidateJSONWithOptions` logs; the other three silently
  swallow bind errors.

Problem:

`context_bind.go:111-133` — `BindAndValidateJSONWithOptions` calls `logBindError`
on both the bind step (line 114) and the validate step (line 128):
```go
func (c *Ctx) BindAndValidateJSONWithOptions(dst any, opts BindOptions) error {
    if err := c.BindJSONWithOptions(dst, opts); err != nil {
        logBindError(c, dst, opts, err)   // ← logs
        return err
    }
    // ...
    if err := validate(dst); err != nil {
        bindErr := &BindError{...}
        logBindError(c, dst, opts, bindErr) // ← logs
        return bindErr
    }
    return nil
}
```

`BindAndValidateJSON` (lines 99-109), `BindAndValidateQuery` (lines 157-165),
and `BindAndValidateQueryWithOptions` (lines 170-185) never call `logBindError`.

This means observability depends on which binding method the handler happens to
call. A caller using `BindAndValidateJSON` (the most common entry point) gets no
automatic bind error logging; a caller using `BindAndValidateJSONWithOptions`
gets it twice.

Fix (preferred): Remove `logBindError` calls from `BindAndValidateJSONWithOptions`
and add a single `logBindError` call to each of the four methods, placed after the
final error is determined (i.e., after validation). This means one log entry per
failed bind/validate call, regardless of which method is used.

Alternative: Remove `logBindError` calls from all four methods and leave logging
entirely to the caller. This is acceptable if the observability policy for bind
errors is documented explicitly.

Whichever direction is chosen, the behavior must be the same across all four
methods.

Scope:
- `BindAndValidateJSON`
- `BindAndValidateJSONWithOptions`
- `BindAndValidateQuery`
- `BindAndValidateQueryWithOptions`

Non-goals:
- Do not change `BindJSON`, `BindJSONWithOptions`, `BindQuery` (non-validate
  variants are caller-logged).
- Do not change `logBindError` internals.

Files:
- `contract/context_bind.go`

Tests:
- Add or update a test that verifies logging behavior is consistent between
  `BindAndValidateJSON` and `BindAndValidateJSONWithOptions`.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- All four `BindAndValidate*` methods have identical logging behavior for
  bind and validation errors.
- No method logs more than once per call.
- All tests pass.

Outcome:
- Completed in the 2026-04-05 contract cleanup batch.
- Verified as part of the shared contract/task-card completion pass.

Validation Run:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
- `go build ./...`
- `go test -timeout 20s ./...`
- `go test -race -timeout 60s ./...`
- `go vet ./...`
