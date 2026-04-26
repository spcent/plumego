# x/ai/tool Beta Evidence

Module: `x/ai/tool`

Owner: `ai-gateway`

Family status: `experimental`

Subpackage tier: `stable`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Registry tests cover register, get, list, execute, duplicate registration,
  and missing tool behavior.
- Policy coverage exercises allow-list filtering and invocation boundaries.
- Builtin tool tests cover echo, calculator, timestamp, read-file, write-file,
  and bash tool behavior.
- Error-result and metrics paths are covered for failed invocations.
- Runnable offline examples demonstrate local tool registration, policy
  filtering, and execution without provider calls.

## Primer And Boundary State

- Primer: `docs/modules/x-ai/README.md`
- Manifest: `x/ai/module.yaml`
- Boundary state: documented as local tool registration and execution policy,
  with no implicit package-global registry.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/ai/tool` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

Missing. Generate snapshots with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/ai/tool -out /tmp/plumego-x-ai-tool-api.snapshot
```

Snapshot refs:

- none recorded

## Owner Sign-Off

Missing. The `ai-gateway` owner must confirm the subpackage beta criteria before
any manifest or dashboard status change.

## Experimental Exclusions

This record does not cover provider tool-call interpretation, orchestration
tool planning, marketplace tool packages, or service-specific tool catalogs.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote the root `x/ai` family. Treat `x/ai/tool` as stable-tier only,
with beta promotion still blocked by missing release and snapshot evidence.
