# Card 1374

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: done
Primary Module: x/websocket
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/modules/x-websocket/README.md`
- `tasks/cards/active/1366-x-websocket-beta-evidence-closure.md`
Depends On: 0732, 0733, 0734, 0735, 0736, 0737, 0739-0772, 0724, release refs and owner sign-off

Goal:
- Close WebSocket maturity evidence only after the runtime stable-readiness cards and release governance evidence are complete.

Problem:
`x/websocket` had runtime stable-readiness hardening completed before its
release governance evidence was available. The card closed after both runtime
cleanup and beta governance evidence were complete.

Scope:
- Keep `x/websocket` out of beta until runtime blockers from cards 0761-0772
  and evidence blockers are closed.
- After release refs exist, generate release-to-release API snapshots.
- Record owner sign-off from `realtime`.
- Update the module README to reflect the final maturity state and stable
  caveats.
- Cross-check that card 0724 remains aligned with any new evidence process.

Non-goals:
- Do not fabricate release refs, snapshots, or owner sign-off.
- Do not promote `x/websocket` based on `HEAD` snapshots alone.
- Do not change runtime behavior in this card.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/modules/x-websocket/README.md`
- `tasks/cards/active/1366-x-websocket-beta-evidence-closure.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required when evidence or maturity state changes.

Done Definition:
- Runtime stable-readiness cards are complete.
- `x/websocket` has real release refs, matching snapshots, and owner sign-off.
- Evidence checks and maturity checks pass without special baselines.
- Promotion state is documented accurately.

Outcome:
- Runtime stable-readiness cards 0739-0772 are complete as of 2026-05-02.
- Release refs `d2c25c3` and `ec70358`, release-backed API snapshots, and
  `realtime` owner sign-off are recorded.
- `x/websocket` is promoted to beta, not GA; future GA promotion still follows
  `docs/EXTENSION_STABILITY_POLICY.md`.

Validation:
- go run ./internal/checks/extension-beta-evidence
- go run ./internal/checks/extension-maturity
- go run ./internal/checks/module-manifests
