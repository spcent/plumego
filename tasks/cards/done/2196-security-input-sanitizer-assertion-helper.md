# Card 2196: Security Input Sanitizer Assertion Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: security/input
Owned Files:
- `security/input/input_test.go`
Depends On: none

Goal:
Consolidate sanitizer contains/not-contains assertions behind local helpers.

Problem:
HTML and SQL sanitizer tests repeat `strings.Contains` checks with nearly
identical error formatting.

Scope:
- Add local helpers for expected and forbidden substrings.
- Use them in sanitizer tests.

Non-goals:
- Do not change sanitizer behavior or public APIs.
- Do not add dependencies.

Files:
- `security/input/input_test.go`

Tests:
- `go test -race -timeout 60s ./security/input/...`
- `go test -timeout 20s ./security/input/...`
- `go vet ./security/input/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- Sanitizer substring assertions use shared helpers.
- The listed validation commands pass.

Outcome:
