# Card 2033

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/docs/design
Priority: P0
State: done
Primary Module: reference/workerfleet/docs/design
Owned Files:
- `reference/workerfleet/docs/design/technical-design.zh-CN.md`
- `reference/workerfleet/README.md`
- `env.example`
Depends On:
- `tasks/cards/done/2031-workerfleet-config-profile-and-policy-wiring.md`
- `tasks/cards/done/2032-workerfleet-metrics-contract-stabilization.md`

## Goal

Bring the Chinese design doc and shared env example back in sync with the current workerfleet runtime, policy, and metrics contract.

## Scope

- Sync the Chinese technical design with profile and policy wiring plus stable and experimental metric language.
- Document the current workerfleet env surface in the shared `env.example`.
- Keep the English and Chinese design docs aligned on implemented behavior.

## Non-goals

- Do not change workerfleet runtime behavior in this card.
- Do not add deployment manifests in this card.
- Do not add new environment variables in this card.

## Files

- `reference/workerfleet/docs/design/technical-design.zh-CN.md`
- `reference/workerfleet/README.md`
- `env.example`

## Acceptance Tests

<!-- Docs-only card: no acceptance test function. -->

## Tests

- Verify the new env keys already exist in app config and are documented consistently.

## Docs Sync

- `reference/workerfleet/docs/design/technical-design.zh-CN.md`
- `reference/workerfleet/README.md`
- `env.example`

## Validation

- `git diff --check`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Synced the Chinese workerfleet design doc with the current runtime shell, profile-driven status and alert policy wiring, and the stable versus experimental metrics contract. Also expanded the shared repository `env.example` so the current workerfleet environment surface is documented in one place for local bootstrapping and operations handoff.
