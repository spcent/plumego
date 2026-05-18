# Card 1373

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: active
Primary Module: x/messaging
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-messaging.md`
- `docs/extension-evidence/release-artifacts.md`
Depends On: release-backed snapshots and owner sign-off

Goal:
- Complete beta evidence closure for the `x/messaging` app-facing service surface.
- This card is part of the post-v1 maturity roadmap recorded in `tasks/cards/active/1452-post-v1-maturity-roadmap.md`.

Problem:
The evidence ledger tracks `x/messaging:app-facing-service` with a first
`v1.0.0` release ref and checked-in baseline snapshot, but it still lacks a
second release ref, complete release-backed snapshots, and owner sign-off.

Scope:
- Add the second real release ref only after the next qualifying tag or release
  commit exists.
- Generate release-to-release API snapshots with `extension-release-evidence`.
- Record owner sign-off from `messaging`.

Non-goals:
- Do not promote subordinate `x/messaging/mq`, `x/messaging/pubsub`, `x/messaging/scheduler`, or `x/messaging/webhook` primitives from this card.
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
