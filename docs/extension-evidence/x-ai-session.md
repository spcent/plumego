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

Missing. Promotion requires two consecutive minor release refs with no exported
`x/ai/session` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

Missing. Generate snapshots with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/ai/session -out /tmp/plumego-x-ai-session-api.snapshot
```

Snapshot refs:

- none recorded

## Owner Sign-Off

Missing. The `ai-gateway` owner must confirm the subpackage beta criteria before
any manifest or dashboard status change.

## Experimental Exclusions

This record does not cover orchestration state machines, semantic cache state,
distributed execution persistence, or any session behavior owned by another
extension family.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote the root `x/ai` family. Treat `x/ai/session` as stable-tier only,
with beta promotion still blocked by missing release and snapshot evidence.
