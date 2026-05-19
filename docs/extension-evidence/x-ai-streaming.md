# x/ai/streaming Beta Evidence

Module: `x/ai/streaming`

Owner: `ai-gateway`

Family status: `experimental`

Subpackage tier: `stable`

Candidate status: `beta`

Evidence state: complete

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

Recorded. Promotion evidence uses two consecutive minor release refs with no
exported `x/ai/streaming` API changes.

Release refs:

- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target)
- `v1.1.0`

## API Snapshot Evidence

Release-backed API snapshots are recorded for the promotion pair below. The
current-head baseline snapshot remains useful during development, but the
release-backed comparison is the promotion evidence.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/ai/streaming -out /tmp/plumego-x-ai-streaming-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/ai-stable-tier/x-ai-streaming-head.snapshot`
- `docs/extension-evidence/snapshots/x-ai-streaming/base.snapshot`
- `docs/extension-evidence/snapshots/x-ai-streaming/head.snapshot`

v1 baseline intake artifacts:

- `docs/extension-evidence/snapshots/v1-baseline/x-ai-streaming/base.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-ai-streaming/head.snapshot`

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/ai/streaming \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-ai-streaming-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until two
release refs and release-backed snapshot evidence are recorded.

## Release Evidence

Release refs: `v1.0.0`, `v1.1.0`

API snapshot comparison:

- Base: `docs/extension-evidence/snapshots/x-ai-streaming/base.snapshot`
- Head: `docs/extension-evidence/snapshots/x-ai-streaming/head.snapshot`
- Result: **API unchanged** across both refs

## Owner Sign-Off

Signed off by `ai-gateway` for v1.1.0:

> I confirm that the `x/ai/streaming` stable-tier subpackage meets the beta
> criteria in docs/EXTENSION_STABILITY_POLICY.md and accept the beta
> compatibility obligations for the documented public surface.

## Scope Exclusions

This record does not cover `x/ai/sse`, orchestration callbacks beyond the
streaming handler contract, distributed execution streams, or live-provider
stream semantics.

## Blockers

None. All promotion blockers cleared.

## Promotion Posture

Do not promote the root `x/ai` family. Treat `x/ai/streaming` as a beta-ready
stable-tier subpackage only.
