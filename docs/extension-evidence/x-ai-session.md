# x/ai/session Beta Evidence

Module: `x/ai/session`

Owner: `ai-gateway`

Family status: `experimental`

Subpackage tier: `stable`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Session manager tests cover create, get, update, delete, list, and active
  session lookup behavior.
- In-memory storage coverage includes TTL expiry, pagination, and message
  append flows.
- Message lifecycle coverage exercises token counting, auto-trim behavior, and
  context key/value storage.
- Runnable offline examples demonstrate in-memory session lifecycle and active
  context retrieval without external services.

## Primer And Boundary State

- Primer: `docs/modules/x-ai/README.md`
- Manifest: `x/ai/module.yaml`
- Boundary state: documented as explicit session ownership in the consuming
  service, not a hidden global session manager.

## Required Release Evidence

Partially recorded. Promotion requires two consecutive minor release refs with
no exported `x/ai/session` API changes. The `v1.0.0` tag target is the first
post-v1 release-ref intake point only; it does not clear
`release_history_missing` by itself.

Release refs:

- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target)

## API Snapshot Evidence

One current-head baseline snapshot is recorded. It is useful for comparing the
subpackage surface during development, but it is not release evidence and does
not clear `api_snapshot_missing` by itself.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/ai/session -out /tmp/plumego-x-ai-session-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/ai-stable-tier/x-ai-session-head.snapshot`

v1 baseline intake artifacts:

- `docs/extension-evidence/snapshots/v1-baseline/x-ai-session/base.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-ai-session/head.snapshot`

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/ai/session \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-ai-session-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until two
release refs and release-backed snapshot evidence are recorded.

## Release Evidence

First release-ref intake recorded.

Release refs: `v1.0.0`

API snapshot comparison: `v1.0.0` to `v1.0.0`, unchanged

## Owner Sign-Off

Missing. The `ai-gateway` owner must confirm the subpackage beta criteria before
any manifest or dashboard status change.

## Scope Exclusions

This record does not cover orchestration state machines, semantic cache state,
distributed execution persistence, or any session behavior owned by another
extension family.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Posture

Do not promote the root `x/ai` family. Treat `x/ai/session` as stable-tier only,
with beta promotion still blocked by missing release and snapshot evidence.
