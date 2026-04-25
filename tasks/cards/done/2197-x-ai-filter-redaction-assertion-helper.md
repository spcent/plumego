# Card 2197: x/ai Filter Redaction Assertion Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: x/ai/filter
Owned Files:
- `x/ai/filter/filter_test.go`
Depends On: none

Goal:
Consolidate redaction substring assertions behind local helpers.

Problem:
Redaction tests repeat `strings.Contains` checks for forbidden original content
and required redaction markers.

Scope:
- Add local contains/not-contains assertion helpers.
- Use them in redaction tests.

Non-goals:
- Do not change filter behavior or public APIs.
- Do not add dependencies.

Files:
- `x/ai/filter/filter_test.go`

Tests:
- `go test -race -timeout 60s ./x/ai/filter/...`
- `go test -timeout 20s ./x/ai/filter/...`
- `go vet ./x/ai/filter/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- Redaction tests use shared substring helpers.
- The listed validation commands pass.

Outcome:
