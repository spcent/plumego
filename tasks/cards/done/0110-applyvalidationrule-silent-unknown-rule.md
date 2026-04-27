# Card 0110

Priority: P2

Goal:
- Make `applyValidationRule` return an error on unknown rule names so that tag
  typos like `validate:"requried"` surface at runtime instead of silently
  passing validation.

Problem:

`validation.go:106-108`:
```go
switch name {
// ...
default:
    return nil   // ← unknown rule silently passes
}
```

If a struct tag contains a misspelled or unsupported rule name, validation
silently succeeds. There is no warning, no error, and no FieldError:

```go
type User struct {
    Name string `validate:"requried"`  // typo — passes as if no rule
    Age  int    `validate:"minn=18"`   // typo — passes as if no rule
}
```

`validateStruct(&User{})` returns nil for both fields. The "required" and "min"
constraints are effectively dropped with zero indication.

Fix: Return a configuration `FieldError` for unknown rule names:
```go
default:
    return &FieldError{
        Field:   fieldName,
        Code:    "unknown_rule",
        Message: fmt.Sprintf("unknown validation rule: %q", name),
    }
```

This converts a silent misconfiguration into a visible validation failure.
Since it produces a `FieldError`, it appears in the `400` response alongside
normal validation failures, making it immediately discoverable during
development without requiring a schema-checking tool.

Note: if callers are already using custom or extended rule names (e.g.,
integrated with a third-party validator via `BindOptions.Validator`), those
callers bypass `applyValidationRule` entirely and are not affected.

Non-goals:
- Do not add runtime rule registration.
- Do not change the rule parsing logic.
- Do not change how `BindOptions.Validator` interacts with validation.

Files:
- `contract/validation.go`

Tests:
- Add a test: a struct with `validate:"unknownrule"` must return a FieldError
  with `Code: "unknown_rule"`.
- Add a test: a struct with `validate:"requried"` (typo) must return a FieldError.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- Unknown rule names produce a `FieldError` instead of silently returning nil.
- Existing tests for known rules still pass.
- New tests for unknown rules pass.

Outcome:
- Completed in the 2026-04-05 contract cleanup batch.
- Verified as part of the shared contract/task-card completion pass.

Validation Run:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
- `go build ./...`
- `go test -timeout 20s ./...`
- `go test -race -timeout 60s ./...`
- `go vet ./...`
