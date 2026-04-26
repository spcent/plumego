# x/ai/provider Beta Evidence

Module: `x/ai/provider`

Owner: `ai-gateway`

Family status: `experimental`

Subpackage tier: `stable`

Candidate status: `beta`

Evidence state: incomplete

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

- Primer: `docs/modules/x-ai/README.md`
- Manifest: `x/ai/module.yaml`
- Boundary state: documented as a stable-tier AI subpackage with explicit
  provider selection and no package-global provider registry.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/ai/provider` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

Missing. Generate snapshots with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/ai/provider -out /tmp/plumego-x-ai-provider-api.snapshot
```

Snapshot refs:

- none recorded

## Owner Sign-Off

Missing. The `ai-gateway` owner must confirm the subpackage beta criteria before
any manifest or dashboard status change.

## Experimental Exclusions

This record does not cover `x/ai/orchestration`, `x/ai/semanticcache`,
`x/ai/marketplace`, `x/ai/distributed`, `x/ai/resilience`, provider-specific
runtime credentials, or live-provider integration behavior.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote the root `x/ai` family. Treat `x/ai/provider` as stable-tier
only, with beta promotion still blocked by missing release and snapshot
evidence.
