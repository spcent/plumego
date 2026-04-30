# Card 0731

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: blocked
Primary Module: x/messaging
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-messaging.md`
- `docs/extension-evidence/release-artifacts.md`
Depends On: release refs and owner sign-off

Goal:
- Complete beta evidence closure for the `x/messaging` app-facing service surface.

Problem:
The evidence ledger tracks `x/messaging:app-facing-service`, but it currently lacks checked-in API snapshots, release history, release snapshots, and owner sign-off.

Scope:
- Generate a current-head snapshot for the app-facing messaging service surface when ready for candidate review.
- Add two real release refs only after tags or release commits exist.
- Generate release-to-release API snapshots with `extension-release-evidence`.
- Record owner sign-off from `messaging`.

Non-goals:
- Do not promote subordinate `x/mq`, `x/pubsub`, `x/scheduler`, or `x/webhook` primitives from this card.
- Do not use `HEAD` as release-history evidence.
- Do not change messaging runtime behavior.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-messaging.md`
- `docs/extension-evidence/release-artifacts.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Required when evidence is added.

Done Definition:
- `x/messaging:app-facing-service` has checked-in snapshots, two release refs, release snapshots, and owner sign-off, or blockers remain explicit.

Outcome:
-
