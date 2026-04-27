# Card 0483: Store File Error String Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: store/file
Owned Files:
- `store/file/coverage_test.go`
Depends On: none

Goal:
Consolidate repeated file error string containment assertions behind a local
helper.

Problem:
`coverage_test.go` repeats `strings.Contains` checks for file error string
parts with slightly different messages.

Scope:
- Add a local helper for asserting an error string contains an expected part.
- Use it in file error string tests.

Non-goals:
- Do not change store/file behavior or public APIs.
- Do not add dependencies.

Files:
- `store/file/coverage_test.go`

Tests:
- `go test -race -timeout 60s ./store/file/...`
- `go test -timeout 20s ./store/file/...`
- `go vet ./store/file/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- File error string tests use a shared containment helper.
- The listed validation commands pass.

Outcome:
