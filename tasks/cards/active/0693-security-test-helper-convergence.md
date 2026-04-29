# Card 0693

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: security
Owned Files:
- security/headers/headers_test.go
- security/input/input_test.go
Depends On: 0692

Goal:
Remove hand-rolled string helpers from security tests and use standard-library helpers consistently.

Scope:
- Replace local `containsDirective`, `splitCSP`, `splitString`, `trimSpace`, and `hasPrefix` helpers with `strings` package calls.
- Replace input test assertion helpers where they only wrap `strings.Contains`.
- Keep test intent and coverage unchanged.

Non-goals:
- Do not change security runtime behavior.
- Do not refactor unrelated tests.

Files:
- `security/headers/headers_test.go`
- `security/input/input_test.go`

Tests:
- `go test -timeout 20s ./security/headers ./security/input`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Not required; test-only cleanup.

Done Definition:
- Security tests no longer maintain duplicate local string parsing helpers for basic contains/split/trim/prefix checks.
- Targeted tests and vet pass.

Outcome:
