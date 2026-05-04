# Card 0720

Milestone: cmd stable hardening
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P0
State: active
Primary Module: cmd/plumego
Owned Files: cmd/plumego/go.mod, cmd/plumego/MODULE.md, cmd/plumego/README.md, docs/release/PRE_V1_RELEASE_CHECKLIST.md
Depends On:

Goal:
Make the CLI release/install story truthful and mechanically verifiable before stable.

Scope:
- Decide and document the supported local-build versus tagged-release install path.
- Remove stale claims that `go install github.com/spcent/plumego/cmd/plumego@vX`
  is verified if the current module shape cannot support it.
- Add an explicit release verification command or fallback install workflow.
- Keep the main module dependency-free.

Non-goals:
- Do not move the CLI into the root module.
- Do not add external dependencies.
- Do not publish a tag or change release metadata outside docs/checklists.

Files:
- `cmd/plumego/go.mod`
- `cmd/plumego/MODULE.md`
- `cmd/plumego/README.md`
- `docs/release/PRE_V1_RELEASE_CHECKLIST.md`

Tests:
- `go build .`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/MODULE.md`
- `cmd/plumego/README.md`
- `docs/release/PRE_V1_RELEASE_CHECKLIST.md`

Done Definition:
- CLI install docs match the repository's actual module layout.
- Release checklist includes a command that would catch broken tagged installs.
- CLI still builds and tests inside `cmd/plumego`.
