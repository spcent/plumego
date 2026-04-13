# Card 0980: x/webhook/out.go Error Builder Redundancy and Freeform Code Cleanup

Priority: P2
State: active
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/webhook
Depends On: 0972

## Goal

`x/webhook/out.go` contains 22 `contract.NewErrorBuilder()` call sites with two
separate defects introduced during the `*contract.Ctx` → `(w, r)` migration in
card 0978:

1. **Redundant `.Status()` before `.Type()`** — `ErrorBuilder.Type()` overwrites
   `Status`, `Category`, and `Code` unconditionally from the type's metadata
   table, so any `.Status(X)` call _before_ `.Type(Y)` is silently discarded.

2. **Redundant `.Category()` after `.Type()`** — `.Type()` already sets
   `Category` from the same metadata table; the trailing `.Category(Z)` call
   just re-sets it to the same value.

3. **Freeform lowercase `Code` strings** — All 22 call sites use a raw
   lowercase string literal (e.g. `"invalid_json"`, `"bad_request"`,
   `"not_found"`, `"store_error"`, `"forbidden"`, `"unauthorized"`) instead of
   the canonical `contract.Code*` constants introduced in card 0972.

## Problem — call order detail

`ErrorBuilder.Type()` is defined as:

```go
func (b *ErrorBuilder) Type(errorType ErrorType) *ErrorBuilder {
    meta := errorType.Meta()
    b.err.Type     = errorType
    b.err.Category = meta.Category   // overwrite
    b.err.Code     = meta.Code       // overwrite
    b.err.Status   = meta.Status     // overwrite
    return b
}
```

So the pattern used throughout `out.go`:

```go
contract.NewErrorBuilder().
    Status(http.StatusBadRequest).      // ← overwritten by .Type()
    Type(contract.TypeValidation).      // ← sets Status=400, Category=Validation, Code=VALIDATION_ERROR
    Code("invalid_json").               // ← overrides code to freeform "invalid_json"
    Message("invalid JSON payload").
    Category(contract.CategoryValidation). // ← redundant; already set by .Type()
    Build()
```

The canonical form following the style guide (§17.3) is:

```go
contract.NewErrorBuilder().
    Type(contract.TypeValidation).
    Code(contract.CodeInvalidJSON).
    Message("invalid JSON payload").
    Build()
```

## Freeform code → canonical constant mapping

| Freeform code | Replace with |
|---|---|
| `"invalid_json"` | `contract.CodeInvalidJSON` |
| `"bad_request"` | `contract.CodeBadRequest` |
| `"not_found"` | `contract.CodeResourceNotFound` |
| `"store_error"` | `contract.CodeInternalError` |
| `"forbidden"` | `contract.CodeForbidden` |
| `"unauthorized"` | `contract.CodeUnauthorized` |

## Scope

All 22 call sites in `x/webhook/out.go`:
- Remove leading `.Status(X)` when it matches what `.Type()` would set anyway
- Remove trailing `.Category(Z)` when it matches what `.Type()` already sets
- Replace all 6 freeform lowercase code strings with the corresponding
  `contract.Code*` constant

## Files

- `x/webhook/out.go`
- `x/webhook/outbound_test.go` (verify test assertions still pass after code changes)

## Tests

```bash
go test -timeout 20s ./x/webhook/...
go vet ./x/webhook/...
```

## Done Definition

- No bare `.Status()` call appears before `.Type()` in the same builder chain
  in `x/webhook/out.go`.
- No bare `.Category()` call appears after `.Type()` in the same builder chain.
- No freeform lowercase string appears in any `.Code("…")` call in
  `x/webhook/out.go`.
- All tests pass; `go vet` clean.

## Outcome

