# x/ai/streaming Beta Evidence

Module: `x/ai/streaming`

Owner: `ai-gateway`

Family status: `experimental`

Subpackage tier: `stable`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Stream manager tests cover registration, lookup, lifecycle, and explicit
  validation via `RegisterE`.
- Streaming engine tests cover progress updates, terminal states, event
  recording, and error propagation.
- Parallel execution tests exercise streaming fan-out behavior and cancellation
  paths.
- HTTP handler tests cover SSE setup, invalid JSON, callback errors, stream
  creation failures, and structured terminal payloads.
- Runnable offline examples demonstrate SSE-backed progress updates with an
  in-memory recorder.

## Primer And Boundary State

- Primer: `docs/modules/x-ai/README.md`
- Manifest: `x/ai/module.yaml`
- Boundary state: documented as streaming coordination for AI services, with
  HTTP error responses using canonical `contract.WriteError` behavior.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/ai/streaming` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

Missing. Generate snapshots with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/ai/streaming -out /tmp/plumego-x-ai-streaming-api.snapshot
```

Snapshot refs:

- none recorded

## Owner Sign-Off

Missing. The `ai-gateway` owner must confirm the subpackage beta criteria before
any manifest or dashboard status change.

## Experimental Exclusions

This record does not cover `x/ai/sse`, orchestration callbacks beyond the
streaming handler contract, distributed execution streams, or live-provider
stream semantics.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote the root `x/ai` family. Treat `x/ai/streaming` as stable-tier
only, with beta promotion still blocked by missing release and snapshot
evidence.
