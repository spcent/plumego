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

Partially recorded. Promotion requires two consecutive minor release refs with
no exported `x/ai/tool` API changes. The `v1.0.0` tag target is the first
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
go run ./internal/checks/extension-api-snapshot -module ./x/ai/tool -out /tmp/plumego-x-ai-tool-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/ai-stable-tier/x-ai-tool-head.snapshot`

v1 baseline intake artifacts:

- `docs/extension-evidence/snapshots/v1-baseline/x-ai-tool/base.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-ai-tool/head.snapshot`

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/ai/tool \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-ai-tool-release-evidence
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

This record does not cover provider tool-call interpretation, orchestration
tool planning, marketplace tool packages, or service-specific tool catalogs.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Posture

Do not promote the root `x/ai` family. Treat `x/ai/tool` as stable-tier only,
with beta promotion still blocked by missing release and snapshot evidence.
