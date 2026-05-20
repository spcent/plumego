# Card 2036

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/docs/design
Priority: P1
State: done
Primary Module: reference/workerfleet/docs/design
Owned Files:
- `reference/workerfleet/docs/design/technical-design.md`
- `reference/workerfleet/docs/design/technical-design.zh-CN.md`
- `reference/workerfleet/docs/alerts.md`
- `reference/workerfleet/README.md`
- `tasks/cards/active/2036-workerfleet-multi-replica-lease-roadmap.md`
Depends On:
- `tasks/cards/done/2027-workerfleet-loop-scheduling-guardrails.md`

## Goal

Document the concrete next-step design for multi-replica loop ownership so workerfleet can evolve from a single active loop runner to safe replicated deployments.

## Scope

- Describe the current single-replica assumption and the risks of duplicate loop execution.
- Define a concrete lease direction, ownership boundaries, and rollout plan.
- Keep the design aligned with the existing loop lease seam in `internal/app`.

## Non-goals

- Do not implement distributed lease code in this card.
- Do not add new runtime config or storage collections in this card.
- Do not change alert or status behavior in code.

## Files

- `reference/workerfleet/docs/design/technical-design.md`
- `reference/workerfleet/docs/design/technical-design.zh-CN.md`
- `reference/workerfleet/docs/alerts.md`
- `reference/workerfleet/README.md`
- `tasks/cards/active/2036-workerfleet-multi-replica-lease-roadmap.md`

## Acceptance Tests

<!-- Docs-only card: no acceptance test function. -->

## Tests

- Keep lease design language specific enough to turn into future implementation cards without re-discovery.

## Docs Sync

- `reference/workerfleet/docs/design/technical-design.md`
- `reference/workerfleet/docs/design/technical-design.zh-CN.md`
- `reference/workerfleet/docs/alerts.md`
- `reference/workerfleet/README.md`

## Validation

- `git diff --check`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Documented the current single-replica constraint for background loops and turned the existing no-op lease seam into a concrete multi-replica roadmap. The design now spells out ownership granularity, the Mongo-backed lease direction, failure and renewal semantics, and the rollout sequence needed before workerfleet replicas can safely scale beyond one active loop owner.
