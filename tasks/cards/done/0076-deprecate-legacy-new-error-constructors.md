# Card 0076

Priority: P1

Goal:
- Mark the 8 legacy `NewXxxError` convenience constructors with proper Go
  `// Deprecated:` doc comments so tooling surfaces the deprecation and callers
  are guided to `NewErrorBuilder()`.

Problem:
- Style guide §16.3 explicitly states:
  > "The 7 existing constructors (NewValidationError, NewNotFoundError, etc.)
  > are legacy; prefer ErrorBuilder in new code. Do not add new NewXxxError(...)
  > convenience constructors for each scenario."
- None of the 8 constructors in `contract/errors.go` carry a `// Deprecated:`
  first-line doc comment, so gopls, staticcheck, and go doc do not surface
  the deprecation to callers.
- The 8 constructors are:
  1. `NewValidationError` (line 362)
  2. `NewNotFoundError` (line 374)
  3. `NewUnauthorizedError` (line 385)
  4. `NewForbiddenError` (line 398)
  5. `NewTimeoutError` (line 411)
  6. `NewInternalError` (line 424)
  7. `NewRateLimitError` (line 437)
  8. `NewBadRequestError` (line 450) — undocumented 8th constructor

Note: card 0502 intended to keep these as the "preferred path for known error
kinds," which contradicts the style guide. This card aligns with the style guide.
The constructors are not removed here (see Done Definition for the boundary).

Scope:
- Add `// Deprecated: Use NewErrorBuilder() instead.` as the first line of
  the doc comment for each of the 8 constructors.
- Add a single block comment above the group (before `NewValidationError`)
  directing readers to `NewErrorBuilder()`.
- Do NOT remove the functions in this card.

Non-goals:
- Do not migrate call sites in this card.
- Do not change function behavior or signatures.
- Removal of the constructors is a follow-up card after call sites are migrated.

Files:
- `contract/errors.go`

Tests:
- `go vet ./contract/...`
- `go test ./contract/...`

Done Definition:
- All 8 constructors have `// Deprecated: Use NewErrorBuilder() instead.` as
  their doc comment first line.
- `go vet` emits a deprecation warning when callers use these functions
  (verified by staticcheck or gopls in CI, if available).
- All tests pass without modification.
