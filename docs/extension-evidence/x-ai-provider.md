# x/ai/provider Beta Evidence

Module: `x/ai/provider`

Owner: `ai-gateway`

Family status: `experimental`

Subpackage tier: `stable`

Candidate status: `beta`

Evidence state: complete

## Current Coverage

- Provider-neutral request, response, usage, tool choice, and streaming types
  are covered by package tests.
- Manager routing coverage exercises registration, default routing,
  Complete/CompleteStream dispatch, and custom router injection.
- Mock provider examples cover runnable offline usage without live provider
  credentials.
- OpenAI and Claude adapter tests use `httptest.NewServer` for offline success
  and API error paths, including option overrides for base URL and HTTP client.

## Primer And Boundary State

- Primer: `docs/modules/x/ai/README.md`
- Manifest: `x/ai/module.yaml`
- Boundary state: documented as a stable-tier AI subpackage with explicit
  provider selection and no package-global provider registry.

## Required Release Evidence

Recorded. Promotion evidence uses two consecutive minor release refs with no
exported `x/ai/provider` API changes.

Release refs:

- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target)
- `v1.1.0`

## API Snapshot Evidence

Release-backed API snapshots are recorded for the promotion pair below. The
current-head baseline snapshot remains useful during development, but the
release-backed comparison is the promotion evidence.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/ai/provider -out /tmp/plumego-x-ai-provider-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/ai-stable-tier/x-ai-provider-head.snapshot`
- `docs/extension-evidence/snapshots/x-ai-provider/base.snapshot`
- `docs/extension-evidence/snapshots/x-ai-provider/head.snapshot`

v1 baseline intake artifacts:

- `docs/extension-evidence/snapshots/v1-baseline/x-ai-provider/base.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-ai-provider/head.snapshot`

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/ai/provider \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-ai-provider-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until two
release refs and release-backed snapshot evidence are recorded.

## Release Evidence

Release refs: `v1.0.0`, `v1.1.0`

API snapshot comparison:

- Base: `docs/extension-evidence/snapshots/x-ai-provider/base.snapshot`
- Head: `docs/extension-evidence/snapshots/x-ai-provider/head.snapshot`
- Result: **API unchanged** across both refs

## Owner Sign-Off

Signed off by `ai-gateway` for v1.1.0:

> I confirm that the `x/ai/provider` stable-tier subpackage meets the beta
> criteria in docs/EXTENSION_STABILITY_POLICY.md and accept the beta
> compatibility obligations for the documented public surface.

## Scope Exclusions

This record does not cover `x/ai/orchestration`, `x/ai/semanticcache`,
`x/ai/marketplace`, `x/ai/distributed`, `x/ai/resilience`, provider-specific
runtime credentials, or live-provider integration behavior.

## Blockers

None. All promotion blockers cleared.

## Promotion Posture

Do not promote the root `x/ai` family. Treat `x/ai/provider` as a beta-ready
stable-tier subpackage only.
