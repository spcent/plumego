# Card 0472: cmd/plumego Codegen NoTODO Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/codegen/codegen_test.go`
Depends On: none

Goal:
Consolidate repeated generated-code `// TODO` assertions behind one local test
helper.

Problem:
`codegen_test.go` repeats the same `strings.Contains(content, "// TODO")`
assertion across middleware, handler, handler-test, and model generation tests.
The repeated checks obscure the actual fixture differences in each case.

Scope:
- Add a local helper for asserting generated content has no bare `// TODO`.
- Replace repeated inline NoTODO assertions with the helper.

Non-goals:
- Do not change generated code.
- Do not change parseability or canonical HTTP contract tests.
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
- Repeated NoTODO assertions use one local helper.
- The listed validation commands pass.

Outcome:
