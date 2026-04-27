# Card 0088

Priority: P2

Goal:
- Fix `validateMin` and `validateMax` for string fields so that `min`/`max`
  always means character count. Remove the hidden numeric-string branch that
  causes surprising validation results.

Problem:

`validation.go:172-183` (`validateMin`, same pattern in `validateMax:213-225`):
```go
case reflect.String:
    text := value.String()
    trimmed := strings.TrimSpace(text)
    if parsed, ok := parseNumericString(trimmed); ok {
        if parsed < float64(limit) {
            return &FieldError{..., Message: fmt.Sprintf("must be at least %d", limit)}
        }
        return nil   // ← exits WITHOUT checking character count
    }
    if int64(utf8.RuneCountInString(text)) < limit {
        return &FieldError{..., Message: fmt.Sprintf("must be at least %d characters", limit)}
    }
```

When the string value parses as a number, the rule is silently interpreted as
a numeric range check instead of a character count check. A struct field
`Amount string \`validate:"min=10"\`` with value `"42"` passes validation
because 42 >= 10 numerically — even though the string is 2 characters long.

Problems:
1. The behavior is invisible from the struct tag. There is no way to know
   whether `min=10` means "10 characters" or "numeric value ≥ 10" without
   reading the source.
2. The error message says "must be at least N" (not "N characters") in the
   numeric path, creating ambiguity in error responses.
3. A caller who wants numeric validation of a string field cannot express
   "only apply length check, not numeric" with the current API.

Fix:
- Remove `parseNumericString` branch from `validateMin` and `validateMax`.
- `min`/`max` on strings always validates UTF-8 rune count.
- If numeric validation of string-encoded values is needed in future, it
  should be a separate rule (e.g., `min_value`, `max_value`) with an explicit
  tag.

Migration: Any struct that relied on the numeric interpretation of `min`/`max`
on string fields must be updated. These are likely few; `grep -rn 'validate:' . --include='*.go'`
to audit.

Non-goals:
- Do not add `min_value`/`max_value` rules in this card.
- Do not change `min`/`max` behavior for int, uint, or float fields.
- Do not change the `required` or `email` rules.

Files:
- `contract/validation.go`
- Any structs with string fields using `min` or `max` validate tags

Tests:
- Add tests: string `"42"` with `min=10` must now FAIL (length 2 < 10).
- Add tests: string `"hello world"` (11 runes) with `max=10` must FAIL.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `parseNumericString` is not called from `validateMin` or `validateMax`.
- `min`/`max` on a string field always validates rune count.
- New behavior is covered by tests.
- All tests pass.

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
