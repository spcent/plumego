# Card 0736

Milestone: cmd stable hardening
Recipe: specs/change-recipes/docs-and-tests.yaml
Priority: P2
State: active
Primary Module: cmd/plumego
Owned Files: cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md, cmd/plumego/MODULE.md, docs/release/PRE_V1_RELEASE_CHECKLIST.md
Depends On: 0735

Goal:
Add stable CLI smoke coverage and final docs sync for the hardened command surface.

Scope:
- Add black-box style tests for `new -> build/check/test` on a generated canonical project where feasible.
- Add output format smoke for JSON/YAML/text help and version behavior.
- Sync release checklist with the stable command smoke set.

Non-goals:
- Do not run network or tagged release installation from tests.
- Do not add slow end-to-end tests that require a long-lived dev server.
- Do not broaden scope outside `cmd/plumego` docs and tests.

Files:
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`
- `cmd/plumego/MODULE.md`
- `docs/release/PRE_V1_RELEASE_CHECKLIST.md`

Tests:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/README.md`
- `cmd/plumego/MODULE.md`
- `docs/release/PRE_V1_RELEASE_CHECKLIST.md`

Done Definition:
- Stable CLI smoke coverage exists for generated project workflows and output formats.
- Release checklist references the smoke set.
- Full cmd module tests and vet pass.
