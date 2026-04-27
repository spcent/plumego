# Card 0063

Priority: P0

Goal:
- Reduce the five parallel paths for constructing an `APIError` to two:
  convenience constructors for common cases, and a single builder for custom
  cases. Remove the raw struct-literal and BindError conversion paths.

Problem:
- Five construction paths reach the same `APIError` type with no guidance on
  which to prefer:
  1. Direct struct literal: `APIError{Status: 400, Code: "...", ...}`
  2. Builder chain: `NewErrorBuilder().Status(400).Code("...").Build()`
  3. Convenience functions: `NewValidationError(field, msg)`
  4. BindError conversion: `BindErrorToAPIError(err)`
  5. ErrorChain conversion: `ErrorHandler.ToAPIError(errorChain)`
- Callers mix these freely, producing inconsistently populated `APIError`
  values (e.g. missing `Category`, missing `Code`).

Scope:
- Designate `NewErrorBuilder()` as the single custom-construction path.
- Keep and expand the convenience functions (`NewValidationError`,
  `NewNotFoundError`, etc.) as the preferred path for known error kinds.
- Funnel `BindErrorToAPIError` and `ErrorHandler.ToAPIError` through the
  builder internally so their output is always fully populated.
- Add a linter note (comment in `errors.go`) discouraging direct struct
  literals outside the `contract` package itself.

Non-goals:
- Do not change `APIError` wire format.
- Do not remove `ErrorBuilder` — it is the designated custom path.
- Do not change `ErrorChain` or `ErrorMetrics` behaviour.

Files:
- `contract/errors.go`
- `contract/error_utils.go`
- `contract/bind_helpers.go`

Tests:
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `BindErrorToAPIError` and `ErrorHandler.ToAPIError` produce fully populated
  `APIError` values (Status, Code, Category all set).
- No call site outside `contract/` constructs `APIError` via direct struct
  literal.
- All tests pass.
