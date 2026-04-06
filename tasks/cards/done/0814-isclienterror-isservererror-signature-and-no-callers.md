# Card 0814

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/errors.go`
- `contract/error_wrap.go`
Depends On: —

Goal:
- Align the signatures of `IsClientError` and `IsServerError` with `IsRetryable`
  so all three error classification functions accept the `error` interface, not
  just the `APIError` concrete type.

Problem:
The package exports three error classification functions with inconsistent signatures:

```go
// errors.go
func IsClientError(err APIError) bool      // accepts concrete type only
func IsServerError(err APIError) bool      // accepts concrete type only

// error_wrap.go
func IsRetryable(err error) bool           // accepts any error, unwraps chain
```

`IsRetryable` accepts `error` and handles `WrappedErrorWithContext` transparently
by walking the unwrap chain. `IsClientError` and `IsServerError` require the
caller to already hold a concrete `APIError`, meaning they must type-assert first:

```go
// Required today — verbose and error-prone
if apiErr, ok := err.(contract.APIError); ok && contract.IsClientError(apiErr) { ... }

// How IsRetryable works — clean
if contract.IsRetryable(err) { ... }
```

Additionally, **neither `IsClientError` nor `IsServerError` has a single external
caller** in the entire codebase (verified by grep). They are exported functions
that no production code uses, and the `APIError` parameter type makes them
awkward to call when working with the `error` interface.

Fix:
- Change both functions to accept `error`:
  ```go
  func IsClientError(err error) bool
  func IsServerError(err error) bool
  ```
- Internally, type-assert to `APIError` (and unwrap `WrappedErrorWithContext`
  if needed, matching `IsRetryable` style).
- If no `APIError` is found in the chain, return `false`.
- Update `errors_test.go` call sites (tests currently pass an `APIError` literal;
  `APIError` implements `error` so no test changes needed except removing explicit
  `APIError{}` type where an `error` variable would be cleaner).

Non-goals:
- No change to the classification logic (4xx = client, 5xx = server).
- No removal of the functions.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `IsClientError(error)` and `IsServerError(error)` accept the `error` interface.
- Passing a plain `errors.New("x")` returns `false` (no APIError in chain).
- Passing an `APIError{Status: 404}` directly or wrapped in
  `WrappedErrorWithContext` returns `true` for `IsClientError`.
- All tests pass.

Outcome:
