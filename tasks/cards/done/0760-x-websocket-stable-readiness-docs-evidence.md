# Card 0760

Milestone: M-003
Recipe: specs/change-recipes/doc-sync.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`
- `x/websocket/doc.go`
- `x/websocket/module.yaml`
- `tasks/cards/active/0738-x-websocket-stable-evidence-readiness.md`
Depends On: 0754, 0755, 0756, 0757, 0758, 0759, 0724

Goal:
- Reconcile docs, module metadata, and evidence after the second WebSocket stable-hardening pass.

Problem:
Runtime hardening and API cleanup need to be reflected in the module README, package examples, and evidence ledger, while still keeping promotion blocked on real release evidence and owner sign-off.

Scope:
- Update package examples so they compile against the final API and use valid secrets.
- Update module docs with final security, lifecycle, protocol, config, and API-surface contracts.
- Update evidence docs to distinguish completed runtime readiness from remaining governance blockers.
- Keep `module.yaml` experimental until real release refs, snapshots, and owner sign-off exist.
- Update card 0738 dependency notes if runtime blockers are complete.

Non-goals:
- Do not fabricate release refs, release snapshots, or owner sign-off.
- Do not promote `x/websocket` to beta or stable.
- Do not widen into other extension modules.

Files:
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`
- `x/websocket/doc.go`
- `x/websocket/module.yaml`
- `tasks/cards/active/0738-x-websocket-stable-evidence-readiness.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
- This is the docs/evidence sync card.

Done Definition:
- Docs describe only implemented behavior.
- Evidence ledger lists remaining governance blockers only.
- Module status remains experimental until governance evidence exists.

Outcome:
- Updated package examples and website guide snippets to use `NewHubWithConfigE`.
- Updated evidence docs to record runtime hardening through cards 0739-0760.
- Updated card 0738 so runtime blockers are complete and only release governance evidence remains.
- Kept `x/websocket` marked experimental; no release refs, release snapshots, or owner sign-off were fabricated.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go run ./internal/checks/extension-beta-evidence`
  - `go run ./internal/checks/module-manifests`
  - `go build ./...`
