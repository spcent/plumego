# Card 0747

Priority: P2
State: active
Primary Module: contract
Owned Files: contract/errors.go, middleware/auth/contract.go

Goal:
- Document that `ErrorBuilder.Type()` overwrites `Status`, `Category`, and
  `Code` with the type's canonical meta values, and remove redundant
  pre-Type calls at the affected call site in `middleware/auth/contract.go`.

Problem:

`errors.go:306-316`:
```go
func (b *ErrorBuilder) Type(errorType ErrorType) *ErrorBuilder {
    if errorType == "" {
        return b
    }
    meta := errorType.Meta()
    b.err.Type = errorType
    b.err.Category = meta.Category   // ← overwrites whatever was set before
    b.err.Code = meta.Code           // ← overwrites whatever was set before
    b.err.Status = meta.Status       // ← overwrites whatever was set before
    return b
}
```

`.Type()` is the only builder method that writes to three fields simultaneously.
Every other setter modifies exactly one field. This makes `.Type()` order-dependent
in a way callers do not expect from a fluent builder:

- Calls to `.Status()`, `.Code()`, or `.Category()` **before** `.Type()` are
  silently overwritten — those calls have no effect.
- Calls to `.Status()`, `.Code()`, or `.Category()` **after** `.Type()` override
  the type defaults — they do take effect.

Neither ordering is documented.

**Effect at the call site in `middleware/auth/contract.go:116-122`:**
```go
contract.NewErrorBuilder().
    Status(http.StatusInternalServerError).    // ← overwritten by Type() below
    Category(contract.CategoryServer).          // ← overwritten by Type() below
    Type(contract.ErrTypeInternal).             // sets Status=500, Category=Server, Code=CodeInternalError
    Code(contract.CodeInternalError).           // redundant: same value Type() just set
    Message(message).
    Build()
```

The three calls before `.Type()` are dead weight. The `.Code()` call after
`.Type()` sets the same value the type already wrote. This was written this way
because the author was unsure which calls would win. The resulting code is
misleading — it looks like all five builder steps contribute, but three are no-ops.

Fix:

1. Add a doc comment to `ErrorBuilder.Type()` explaining the semantics:
   ```go
   // Type sets the error type and populates Category, Code, and Status with the
   // canonical values for that type (from errorTypeLookup). Any Category, Code,
   // or Status set before calling Type will be overwritten. To customize these
   // fields beyond the type defaults, call Category/Code/Status after Type.
   func (b *ErrorBuilder) Type(errorType ErrorType) *ErrorBuilder {
   ```

2. Clean up the redundant builder chain in `middleware/auth/contract.go`:
   ```go
   // Before (redundant):
   contract.NewErrorBuilder().
       Status(http.StatusInternalServerError).
       Category(contract.CategoryServer).
       Type(contract.ErrTypeInternal).
       Code(contract.CodeInternalError).
       Message(message).
       Build()

   // After (minimal):
   contract.NewErrorBuilder().
       Type(contract.ErrTypeInternal).
       Message(message).
       Build()
   ```

Non-goals:
- Do not change how `.Type()` works.
- Do not audit every builder call site in the codebase; fix only `auth/contract.go`
  (the clearest example of the confusion).
- Do not change `errorTypeLookup` or `ErrorType.Meta()`.

Files:
- `contract/errors.go`
- `middleware/auth/contract.go`

Tests:
- Add a test in `contract/errors_test.go` asserting that calling `.Status()` or
  `.Code()` before `.Type()` does not affect the final APIError values:
  ```go
  func TestBuilderTypeOverwritesPriorFields(t *testing.T) {
      got := NewErrorBuilder().
          Status(999).
          Code("CUSTOM").
          Type(ErrTypeNotFound).
          Build()
      if got.Status != http.StatusNotFound {
          t.Errorf("Status: want %d, got %d", http.StatusNotFound, got.Status)
      }
      if got.Code != CodeResourceNotFound {
          t.Errorf("Code: want %q, got %q", CodeResourceNotFound, got.Code)
      }
  }

  func TestBuilderStatusAfterTypeWins(t *testing.T) {
      got := NewErrorBuilder().
          Type(ErrTypeNotFound).
          Status(422).
          Build()
      if got.Status != 422 {
          t.Errorf("Status after Type: want 422, got %d", got.Status)
      }
  }
  ```
- `go test ./contract/...`
- `go test ./middleware/auth/...`
- `go vet ./...`

Done Definition:
- `ErrorBuilder.Type()` has a doc comment explaining its multi-field overwrite
  behaviour and the correct ordering.
- The redundant pre-`Type` calls are removed from `middleware/auth/contract.go`.
- New tests covering both orderings pass.
- All existing tests pass.

Outcome:
- Completed by documenting the overwrite behavior of `ErrorBuilder.Type()`,
  removing redundant pre-`Type` builder calls from `middleware/auth`, and
  adding tests for both “before Type gets overwritten” and “after Type wins”.

Validation Run:
- `gofmt -w contract/errors.go contract/errors_test.go middleware/auth/contract.go`
- `go test -timeout 20s ./contract/... ./middleware/auth/...`
- `go vet ./...`
