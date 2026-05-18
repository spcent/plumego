# x/ai/tool Beta Evidence

Module: `x/ai/tool`

Owner: `ai-gateway`

Family status: `experimental`

Subpackage tier: `stable`

Candidate status: `beta`

Evidence state: complete

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

Recorded. Promotion evidence uses two consecutive minor release refs with no
exported `x/ai/tool` API changes.

Release refs:

- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target)
- `v1.1.0`

## API Snapshot Evidence

Release-backed API snapshots are recorded for the promotion pair below. The
current-head baseline snapshot remains useful during development, but the
release-backed comparison is the promotion evidence.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/ai/tool -out /tmp/plumego-x-ai-tool-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/ai-stable-tier/x-ai-tool-head.snapshot`
- `docs/extension-evidence/snapshots/x-ai-tool/base.snapshot`
- `docs/extension-evidence/snapshots/x-ai-tool/head.snapshot`

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

Release refs: `v1.0.0`, `v1.1.0`

API snapshot comparison:

- Base: `docs/extension-evidence/snapshots/x-ai-tool/base.snapshot`
- Head: `docs/extension-evidence/snapshots/x-ai-tool/head.snapshot`
- Result: **API unchanged** across both refs

## Owner Sign-Off

Signed off by `ai-gateway` for v1.1.0:

> I confirm that the `x/ai/tool` stable-tier subpackage meets the beta criteria
> in docs/EXTENSION_STABILITY_POLICY.md and accept the beta compatibility
> obligations for the documented public surface.

## Scope Exclusions

This record does not cover provider tool-call interpretation, orchestration
tool planning, marketplace tool packages, or service-specific tool catalogs.

## Blockers

None. All promotion blockers cleared.

## Promotion Posture

Do not promote the root `x/ai` family. Treat `x/ai/tool` as a beta-ready
stable-tier subpackage only.
