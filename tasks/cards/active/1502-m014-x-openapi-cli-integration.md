# Card 1502

Milestone: M-014
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P2
State: active
Primary Module: x/openapi
Owned Files:
- `docs/modules/x-openapi/README.md`
- `x/openapi/module.yaml`

Goal:
- Update x/openapi/module.yaml to remove the stale "handled by later cards" note
  and reflect the implemented state.
- Evaluate x/openapi for beta promotion based on release-history evidence.

Problem (updated 2026-05-20):
The CLI integration (`plumego generate spec`) and serialization (JSON + YAML via
stdlib) were already fully implemented in `cmd/plumego/commands/spec.go` and
`x/openapi/marshal.go`. The module primer has been expanded in this session.
The remaining work is:
1. Remove the stale placeholder note from x/openapi/module.yaml summary.
2. Collect release-history evidence once a second qualifying release exists.
3. Evaluate for beta promotion using the standard promotion checklist.

Scope:
- Update x/openapi/module.yaml: remove "Serialization and CLI wiring are handled
  by later OpenAPI milestone cards" from the summary field; update to reflect
  current implemented state.
- Run `go run ./internal/checks/module-manifests` after manifest update.
- Once v1.2.0 exists: run extension-release-evidence, record snapshot, get
  owner sign-off, update extension-beta-evidence.yaml.

Non-goals:
- Do not change x/openapi runtime behavior.
- Do not add external OpenAPI validation library dependency.
- Do not promote to beta without two qualifying release refs.

Files:
- `x/openapi/module.yaml`
- `specs/extension-beta-evidence.yaml` (when promotion evidence is recorded)
- `docs/EXTENSION_MATURITY.md` (when promotion status changes)

Tests:
- `go test -race -timeout 60s ./x/openapi/...`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Update `docs/ROADMAP.md` Phase 14 to note CLI completion and current status.

Done Definition:
- x/openapi/module.yaml summary no longer contains stale milestone-card language.
- x/openapi either has beta status in the evidence ledger, OR has explicit
  blockers with the missing release ref identified.

Notes:
- `plumego generate spec --format json|yaml --output path` is the canonical CLI.
- The hints file is `plumego.spec.yaml` in the project root (optional).
- See `docs/modules/x-openapi/README.md` for the full API surface documentation.
