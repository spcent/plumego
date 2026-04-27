# Card 4102: Text Log Field Key Escaping

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: log
Owned Files:
- `log/logger.go`
- `log/logger_semantics_test.go`
- `docs/modules/log/README.md`
Depends On:
- `tasks/cards/done/3102-log-field-semantics.md`

Goal:
Make text field keys unambiguous for punctuation and control characters, not
only whitespace and equals signs.

Problem:
`formatTextFieldKey` currently quotes empty keys and keys containing whitespace
or `=`, but leaves other punctuation such as quotes, brackets, commas, and
colons unquoted. That makes text output less predictable for non-simple field
names while values already get a clearer escaping path.

Scope:
- Define a conservative safe key character set.
- Quote keys outside that set with `strconv.Quote`.
- Add focused text-output tests.
- Update docs if the key rule becomes explicit.

Non-goals:
- Do not reject or drop caller fields.
- Do not add secret redaction.
- Do not change JSON field names.

Tests:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- Unsafe text keys are quoted deterministically.
- Existing simple keys remain unquoted.
- The listed validation commands pass.
