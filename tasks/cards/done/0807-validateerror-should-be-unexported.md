# Card 0807

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/errors.go`
- `contract/errors_test.go`
Depends On: —

Goal:
- Unexport `ValidateError` — it is an internal guard called only by `WriteError`
  and its own test, and its `[]string` return type is inconsistent with every
  other `Validate*` function in the package.

Problem:
The package exports three `Validate*` functions with different return types:

| Function | Input | Returns |
|---|---|---|
| `ValidateError(APIError)` | APIError struct | `[]string` |
| `ValidateStruct(any)` | user struct | `error` |
| `ValidateCtxHandler(CtxHandlerFunc)` | handler func | `error` |

`ValidateError` is the odd one out: it returns `[]string` (a plain list of
issue strings) while the other two return `error`. A caller reading the public API
has to remember different calling conventions for functions with near-identical
names.

Callers outside the package: **none**. Both callers are package-internal:
- `errors.go`: `WriteError` calls it for its pre-write guard
- `errors_test.go`: the test exercises it directly (internal test, `package contract`)

Since all callers are already inside the package, making the function unexported
removes a confusing public API surface with zero migration cost.

Scope:
- Rename `ValidateError` → `validateAPIError` in `contract/errors.go`.
- Update the two call sites in `errors.go` and `errors_test.go`.
- No other callers exist (verify with `grep -rn 'ValidateError' . --include='*.go'`).

Non-goals:
- No change to the function's logic or return type.
- No change to `ValidateStruct` or `ValidateCtxHandler`.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- After renaming, `grep -rn '\bValidateError\b' . --include='*.go'` must return empty.

Docs Sync: —

Done Definition:
- `ValidateError` does not appear in any exported symbol.
- `validateAPIError` is used by `WriteError` and the test.
- All tests pass.

Outcome:
