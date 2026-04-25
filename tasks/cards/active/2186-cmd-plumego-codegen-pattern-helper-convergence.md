# Card 2186: cmd/plumego Codegen Pattern Helper Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/codegen/codegen_test.go`
Depends On: none

Goal:
Make codegen canonical pattern assertions use explicit include/exclude helper
functions.

Problem:
The canonical HTTP contract test repeats pattern loops inline for required and
disallowed snippets. Local helpers make the intent clearer and reduce future
copy/paste if more generator contract tests are added.

Scope:
- Add local helpers for asserting generated content contains or omits patterns.
- Use them in `TestGenerateHandlerCode_UsesCanonicalHTTPContract`.

Non-goals:
- Do not change generated code.
- Do not change pattern lists or expected canonical contract.
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
- Canonical codegen pattern assertions use local include/exclude helpers.
- The listed validation commands pass.

Outcome:
