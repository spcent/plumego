# Card 0062

Priority: P0

Goal:
- Unify the fragmented Write* response API in the `contract` package so callers
  have one clear path for success responses and one for error responses, with a
  consistent return type.

Problem:
- Seven similar functions exist with three different naming prefixes (`Write*`,
  `WriteContract*`, `WriteHTTP*`) and two different return conventions
  (returns `error` vs. void), making the API surface unpredictable.
- Void-returning variants silently discard write errors.

Current surface:
```
WriteJSON(w, status, payload) error          ← returns error
WriteResponse(w, r, status, data, meta) error ← returns error
WriteError(w, r, err APIError)               ← void
WriteContractResponse(ctx, status, data)     ← void, Contract prefix
WriteContractError(ctx, status, code, msg)   ← void, Contract prefix
WriteHTTPResponse(w, r, status, data)        ← void, HTTP prefix
WriteHTTPError(w, r, status, code, msg)      ← void, HTTP prefix
(c *Ctx).JSON(status, data) error
(c *Ctx).Response(status, data, meta) error
(c *Ctx).ErrorJSON(status, code, msg) error
```

Scope:
- Decide canonical forms:
  - `WriteResponse(w, r, status, data, meta) error` → keep as the primary HTTP-level helper
  - `WriteError(w, r, err APIError) error` → add return value
  - `(c *Ctx).Response()` / `(c *Ctx).Error()` → keep as Ctx-level wrappers
- Deprecate or remove `WriteContractResponse`, `WriteContractError`,
  `WriteHTTPResponse`, `WriteHTTPError` — they are aliases with no distinct
  behaviour.
- Ensure all Write* variants return `error`.

Non-goals:
- Do not change the JSON wire format.
- Do not touch stream-related methods in this card.

Files:
- `contract/response.go`
- `contract/context_response.go`
- `contract/context_core.go`
- All files calling the deprecated helpers (search `WriteContractResponse`,
  `WriteHTTPResponse`)

Tests:
- `go test ./contract/...`
- `go vet ./contract/...`
- `go build ./...`

Done Definition:
- At most two exported Write* helpers exist at the package level (one for
  success, one for error), both returning `error`.
- `WriteContract*` and `WriteHTTP*` variants are removed.
- All callers updated.
- All tests pass.
