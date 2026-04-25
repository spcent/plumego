# Card 2181: cmd/plumego Devserver Analyzer Fixture DTO Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/devserver/analyzer_test.go`
Depends On: none

Goal:
Make the analyzer snapshot test serve a typed config response fixture instead
of nested generic maps.

Problem:
`TestGetAppSnapshot` builds the `_debug/config` response with nested
`map[string]any` despite the analyzer decoding into `devtools.ConfigSnapshot`.
The map fixture duplicates the response shape less clearly than the target DTO.

Scope:
- Replace the nested map response fixture with a typed `devtools.ConfigSnapshot`
  envelope.
- Preserve the existing analyzer assertions.

Non-goals:
- Do not change analyzer behavior or public APIs.
- Do not change dashboard/devserver response behavior.
- Do not add dependencies.

Files:
- `cmd/plumego/internal/devserver/analyzer_test.go`

Tests:
- from `cmd/plumego`: `go test -race -timeout 60s ./internal/devserver/...`
- from `cmd/plumego`: `go test -timeout 20s ./internal/devserver/...`
- from `cmd/plumego`: `go vet ./internal/devserver/...`

Docs Sync:
No docs change required; this is test fixture cleanup.

Done Definition:
- The analyzer snapshot test no longer builds the fixed config response through
  nested maps.
- The listed validation commands pass.

Outcome:
