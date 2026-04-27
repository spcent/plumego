# Card 0082

Priority: P2

Goal:
- Remove `BindOptions.Redact` from the binding configuration struct because
  it is a logging concern, not a binding concern, and its presence in
  `BindOptions` violates single-responsibility.

Problem:
- `contract/bind_helpers.go:16`:
  ```go
  type BindOptions struct {
      MaxBodySize           int64
      DisallowUnknownFields bool
      DisableValidation     bool
      Validator             func(any) error
      Redact                func(any) any   // ← only used in logBindError()
  }
  ```
- `Redact` is only used in `logBindError` (context_bind.go:292-295):
  ```go
  if opts.Redact != nil {
      fields["payload"] = opts.Redact(payload)
  } else {
      fields["payload"] = DefaultObservabilityPolicy.RedactFields(...)["payload"]
  }
  ```
- The three fields above it (`MaxBodySize`, `DisallowUnknownFields`, `Validator`)
  are all about *how to decode and validate the request body*. `Redact` is
  about *how to sanitize sensitive data before logging*. These concerns
  belong in different structs.
- When `Redact` is `nil`, `logBindError` already falls back to
  `DefaultObservabilityPolicy.RedactFields`, which is the correct default.
  The `Redact` override is never set by any current caller.

Scope:
- Remove `Redact func(any) any` from `BindOptions`.
- Update `logBindError` to always use `DefaultObservabilityPolicy.RedactFields`
  (which it already does as the fallback). Delete the `if opts.Redact != nil`
  branch.
- Search for any external caller that sets `BindOptions{Redact: ...}` and
  migrate them (expected: zero callers).

Non-goals:
- Do not add a replacement `LoggingOptions` struct in this card.
- Do not change `DefaultObservabilityPolicy` usage in `logBindError`.
- Do not change any other field in `BindOptions`.

Files:
- `contract/bind_helpers.go` (remove field)
- `contract/context_bind.go` (simplify logBindError)
- Any caller that sets `opts.Redact` (run grep first)

Prerequisite:
```bash
grep -rn 'BindOptions{' . --include='*.go'
grep -rn '\.Redact\s*=' . --include='*.go'
```
Confirm zero callers set the Redact field before deleting it.

Tests:
- `go test ./contract/...`
- `go vet ./contract/...`
- `go build ./...`

Done Definition:
- `BindOptions.Redact` field does not exist.
- `logBindError` always uses `DefaultObservabilityPolicy.RedactFields`.
- All tests pass.
