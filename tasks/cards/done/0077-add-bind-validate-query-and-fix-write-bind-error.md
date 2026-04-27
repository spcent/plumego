# Card 0077

Priority: P1

Goal:
- Add `BindAndValidateQuery` to complete bind-validate symmetry for query
  parameters, and fix `WriteBindError` silently discarding its write error.

Problems:

### P1-A: Missing BindAndValidateQuery

`context_bind.go` has:
- `BindJSON` + `BindAndValidateJSON` (symmetric)
- `BindJSONWithOptions` + `BindAndValidateJSONWithOptions` (symmetric)
- `BindQuery` — but NO `BindAndValidateQuery`

Handlers that bind query parameters and need struct-tag validation must
manually call `validateStruct` (unexported) after `BindQuery`, duplicating
the pattern already established for JSON binding.

### P1-B: WriteBindError discards write error

`bind_helpers.go:99`:
```go
func WriteBindError(w http.ResponseWriter, r *http.Request, err error) {
    _ = WriteError(w, r, BindErrorToAPIError(err))
}
```

The return value of `WriteError` is silently discarded. Network write errors
(e.g., client disconnected) are invisible to callers. All other Write* helpers
in the package return `error`.

Scope:
- Add `(c *Ctx) BindAndValidateQuery(dst any) error` that calls `BindQuery`
  then runs `validateStruct(dst)`, returning a `*BindError` on failure.
  Follow the same pattern as `BindAndValidateJSON` (context_bind.go:99-109).
- Change `WriteBindError` signature to return `error`:
  ```go
  func WriteBindError(w http.ResponseWriter, r *http.Request, err error) error {
      return WriteError(w, r, BindErrorToAPIError(err))
  }
  ```
- Update any caller of `WriteBindError` that discards its result to use `_`
  explicitly, or propagate the error where appropriate.

Non-goals:
- Do not add `BindQueryWithOptions`; that is a follow-up if needed.
- Do not change `BindQuery` behavior.
- Do not change `validateStruct` visibility in this card.

Files:
- `contract/context_bind.go`
- `contract/bind_helpers.go`
- All callers of `WriteBindError`

Tests:
- `go test ./contract/...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `(c *Ctx) BindAndValidateQuery(dst any) error` exists and is tested.
- `WriteBindError` returns `error`.
- No callers silently discard `WriteBindError` return without explicit `_`.
- All tests pass.
