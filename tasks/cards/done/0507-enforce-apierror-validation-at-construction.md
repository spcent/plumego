# Card 0507

Priority: P1

Goal:
- Prevent partially populated `APIError` values from reaching `WriteError` by
  enforcing required fields at the point of construction, not as post-hoc
  defaults inside the write path.

Problem:
- `WriteError()` silently fills in defaults for `Status`, `Code`, and
  `Category` when they are zero/empty:
  ```go
  if err.Status == 0 { err.Status = 500 }
  if err.Code == ""  { err.Code = http.StatusText(err.Status) }
  if err.Category == "" { /* infer */ }
  ```
- `ValidateError()` exists but is never called in the write path.
- This means callers can accidentally omit required fields and receive silent
  fallback behaviour rather than a build-time or test-time failure.

Scope:
- Move the default-filling logic from `WriteError()` into the `ErrorBuilder`
  `.Build()` method (or a dedicated `NewAPIError(...)` constructor).
- Make `WriteError()` call `ValidateError()` and log a warning (not panic)
  if validation fails, so that production traffic is not disrupted while
  misconfigured callers are still fixed.
- Add a test that passes a zero-value `APIError{}` to `WriteError()` and
  asserts the warning is emitted.

Non-goals:
- Do not panic in `WriteError()` — the existing fallback is safer than a
  production panic.
- Do not change the `APIError` JSON wire format.

Files:
- `contract/errors.go`
- `contract/response.go`
- `contract/context_response.go`

Tests:
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `ErrorBuilder.Build()` returns a fully populated `APIError` (Status, Code,
  Category all non-zero).
- `WriteError()` calls `ValidateError()` and logs a warning on failure.
- Existing tests still pass; new test for zero-value input passes.
