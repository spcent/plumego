# Card 0307

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/context_bind.go`
Depends On:

Goal:
- Replace the `BindJSON(dst any, opts ...BindOptions)` variadic signature with `BindJSON(dst any, opts *BindOptions)` so the "at most one" constraint is enforced by the compiler rather than by a runtime 400 error.

Problem:
- `BindJSON` accepts a variadic `...BindOptions` parameter (`context_bind.go:19`), which in Go means callers can pass zero, one, or many values.
- `normalizeJSONBindOptions` (`context_bind.go:188`) rejects more than one option at runtime:
  ```go
  default:
      return BindOptions{}, &bindError{
          Status:  http.StatusBadRequest,
          Message: "BindJSON accepts at most one BindOptions value",
      }
  ```
- A caller who accidentally writes `BindJSON(dst, opts1, opts2)` gets a 400 HTTP response sent to the actual client with the message `"BindJSON accepts at most one BindOptions value"`. This surfaces a programmer error as a client-visible bad request, which is both incorrect and confusing.
- The variadic signature signals to callers "you may pass any number of options", but the implementation immediately contradicts this by failing if more than one is passed.
- The idiomatic Go pattern for an optional single-value parameter is a pointer: `opts *BindOptions`. Nil means "use defaults"; non-nil means "use this config". This is statically checked and carries no runtime ambiguity.

Scope:
- Change `BindJSON(dst any, opts ...BindOptions) error` to `BindJSON(dst any, opts *BindOptions) error`.
- Replace `normalizeJSONBindOptions` with a simpler inline nil check:
  ```go
  if opts == nil {
      opts = &BindOptions{}
  }
  ```
- Update all call sites of `BindJSON` in the package (tests, helpers).
- Callers who previously wrote `c.BindJSON(dst, BindOptions{DisallowUnknownFields: true})` now write `c.BindJSON(dst, &BindOptions{DisallowUnknownFields: true})`.
- Callers who wrote `c.BindJSON(dst)` (no options) continue to work unchanged (nil pointer = defaults).

Non-goals:
- Do not change `BindQuery` (it already has no options parameter).
- Do not add more fields to `BindOptions`.

Files:
- `contract/context_bind.go`
- `contract/write_bind_error_test.go`
- `contract/bind_helpers_test.go`
- Any call site of `BindJSON` in the repository.

Tests:
- `go test -timeout 20s ./contract/...`
- `go build ./...`
- `go vet ./contract/...`

Docs Sync:
- None required.

Done Definition:
- `BindJSON` signature is `(dst any, opts *BindOptions) error`.
- `normalizeJSONBindOptions` is removed.
- No call site passes multiple options.
- All tests pass.

Outcome:
- Pending.
