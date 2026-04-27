# Card 0196

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/errors.go`
Depends On: —

Goal:
- Standardise `ErrorType` wire values so all values use the compound
  `{qualifier}_{noun}` pattern, removing the three bare-word outliers.

Problem:
Most `ErrorType` wire values use a compound `word_word` pattern:

```
"validation_error", "required_field_missing", "invalid_format",
"value_out_of_range", "duplicate_value",
"invalid_token", "expired_token",
"resource_not_found", "resource_conflict", "resource_already_exists", "resource_gone",
"internal_error", "service_unavailable", "rate_limited", "maintenance_mode",
"invalid_state", "insufficient_funds", "operation_not_allowed"
```

Three values break the pattern with a bare single word:

| Constant | Current value | Pattern violation |
|---|---|---|
| `ErrTypeTimeout` | `"timeout"` | bare noun; siblings use `"…_error"` / `"…_mode"` |
| `ErrTypeUnauthorized` | `"unauthorized"` | bare adjective; sibling `"invalid_token"` is `adj_noun` |
| `ErrTypeForbidden` | `"forbidden"` | bare adjective; sibling `"invalid_token"` is `adj_noun` |

A reader scanning the error type list must notice the inconsistency and wonder
whether it is intentional. The wire format is also part of the public API surface
clients use to route errors in dashboards and alerts.

Proposed renames (wire value only; constant names stay the same):
- `ErrTypeTimeout` → `"timeout_error"`
  matches the `"…_error"` suffix used by `"internal_error"` and `"validation_error"`
- `ErrTypeUnauthorized` → `"unauthorized_request"`
  matches the `noun_noun` / `adj_noun` pattern; clarifies the subject
- `ErrTypeForbidden` → `"forbidden_request"`
  same rationale as above

Scope:
- Update the three string values in `contract/errors.go`.
- Check for any code that compares against the old wire strings
  (`grep -rn '"timeout"\|"unauthorized"\|"forbidden"' . --include='*.go'`).
  String comparisons against `ErrorType` values (rather than using the constant)
  must be updated.
- Update tests that assert the old wire values.

Non-goals:
- No change to constant names (`ErrTypeTimeout`, `ErrTypeUnauthorized`,
  `ErrTypeForbidden` remain).
- No change to `errorTypeLookup` entries beyond the string value key.
- No change to `ErrorCategory` or `Code*` constants.

Files:
- `contract/errors.go`
- `contract/errors_test.go`
- `contract/error_handling_test.go`
- Any caller that string-compares against the old values

Tests:
- `go build ./...`
- `go test -timeout 20s ./...`
- After change, `grep -rn '"timeout"\b\|"unauthorized"\b\|"forbidden"\b' . --include='*.go'`
  must return only unrelated string uses (non-ErrorType comparisons).

Docs Sync: —

Done Definition:
- The three ErrorType wire values follow the compound pattern.
- No code compares against the old bare wire strings.
- All tests pass.

Outcome:
