# Card 0602

Priority: P1

Goal:
- Make `Ctx.ErrorJSON` produce the same JSON envelope as `WriteError` so all
  error-response paths share one wire format.

Problem:
- `WriteError` (errors.go:164-195) wraps the error in `ErrorResponse`:
  ```go
  return json.NewEncoder(w).Encode(ErrorResponse{Error: err})
  // wire: {"error": {"code":"...","message":"...","category":"..."}}
  ```
- `Ctx.ErrorJSON` (context_response.go:17-31) calls `c.JSON(status, payload)`
  with a bare `APIError`, bypassing the envelope:
  ```go
  return c.JSON(status, payload)
  // wire: {"code":"...","message":"...","category":"..."}
  ```
- A handler that switches from `Ctx.ErrorJSON` to `WriteError` (or vice versa)
  silently changes the JSON shape clients receive, breaking API contracts.
- Style guide §16.4 calls `Ctx.JSON` a low-level primitive "for cases with no
  envelope"; `Ctx.ErrorJSON` is not such a case.

Scope:
- Rewrite `Ctx.ErrorJSON` to call `WriteError(c.W, c.R, payload)` instead of
  `c.JSON(status, payload)`, so it goes through the envelope and trace-ID
  injection.
- Verify that `Ctx.ErrorJSON` respects `WarnFunc` for partially-populated
  errors (inheriting WriteError's validation guard added in card 0507).
- Search for callers of `Ctx.ErrorJSON` and confirm their test assertions match
  the new envelope shape `{"error": {...}}`.

Non-goals:
- Do not change the `APIError` struct fields.
- Do not change `WriteError` itself.
- Do not rename `Ctx.ErrorJSON`.

Files:
- `contract/context_response.go`
- Any handler or test that calls `Ctx.ErrorJSON` and asserts on the response body

Tests:
- `go test ./contract/...`
- `go vet ./contract/...`
- `go build ./...`

Done Definition:
- `Ctx.ErrorJSON` delegates to `WriteError`.
- A test confirms that `Ctx.ErrorJSON` output equals `{"error": {...}}`, not
  `{...}` at top-level.
- All existing tests pass.
