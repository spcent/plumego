# 0767 - Core README Lifecycle Ownership

State: done
Priority: P1
Primary module: core docs

## Goal

Remove top-level README wording that implies core owns env loading or readiness flags.

## Scope

- Update README lifecycle bullets to describe only core-owned lifecycle behavior.
- Keep README and README_CN semantically aligned.
- Preserve existing links and surrounding docs structure.

## Non-goals

- Do not move readiness behavior into core.
- Do not change app-local env loading guidance.
- Do not alter code.

## Files

- `README.md`
- `README_CN.md`
- `tasks/cards/active/0767-core-readme-lifecycle-ownership.md`

## Tests

- `bash scripts/check-doc-snippets-compile.sh`

## Docs Sync

Top-level bilingual README only.

## Done Definition

- Core lifecycle copy no longer lists env loading or ready flags as core features.
- Doc snippet compile check passes.

## Validation

- `rg -n "ready flags|就绪标志|Environment variable loading|环境变量加载" README.md README_CN.md docs/modules/core/README.md core`
- `bash scripts/check-doc-snippets-compile.sh`
