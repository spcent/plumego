# Card 2184: cmd/plumego Codegen Parse Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/codegen/codegen_test.go`
Depends On: none

Goal:
Consolidate repeated generated Go parseability assertions behind one local
helper.

Problem:
`codegen_test.go` repeats `parser.ParseFile` setup and error formatting across
middleware, handler, handler-test, and model generation tests. The duplicated
boilerplate makes the individual generation fixtures harder to scan.

Scope:
- Add a local helper for parsing generated Go source in tests.
- Replace repeated parseability assertions with the helper.

Non-goals:
- Do not change generated code.
- Do not change canonical HTTP contract assertions.
- Do not add dependencies.

Files:
- `cmd/plumego/internal/codegen/codegen_test.go`

Tests:
- from `cmd/plumego`: `go test -race -timeout 60s ./internal/codegen/...`
- from `cmd/plumego`: `go test -timeout 20s ./internal/codegen/...`
- from `cmd/plumego`: `go vet ./internal/codegen/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- Generated-code parseability tests share one parse helper.
- The listed validation commands pass.

Outcome:
