# Plumego Roadmap

## Current Baseline

- stable roots with explicit boundaries: `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`
- extension discovery and task-routing metadata under `specs/*`
- a single canonical application layout in `reference/standard-service`
- milestone, plan, card, and verify workflow assets under `tasks/*`
- repo-wide quality gates in `Makefile` and `.github/workflows/quality-gates.yml`
- stable-root compatibility policy in `docs/reference/deprecation.md`
- stable-root exported API baseline snapshots under `docs/evidence/stable-api/snapshots`
- release evidence checklist under `docs/release/PRE_V1_RELEASE_CHECKLIST.md`
- beta promotion checklist and card template under `docs/release/PROMOTION_CARD_TEMPLATE.md`
- `x/rest`, `x/websocket`, `x/gateway`, and `x/observability` promoted to `beta`
- `v1.0.0` tagged on May 15, 2026; release notes and evidence in `docs/release/v1.0.0.md`

## Phase 16: Active Extension Evaluations

Status: in progress

Evaluate remaining extension promotion candidates against the beta evidence policy. Each candidate requires a second release ref, release-backed API snapshots, and owner sign-off. See `specs/extension-beta-evidence.yaml` for current blocker status per candidate.

## Suggested Execution Order

1. Keep Phase 13 and Phase 15 docs and onboarding sync continuous.
2. Keep `docs/concepts/extension-maturity.md` aligned with module manifests and evidence records.
3. Execute task card 1500 (x/tenant GA evaluation) once v1.2.0 release evidence is available.
4. Execute task card 1501 (x/ai stable-tier subpackages beta evaluation) per-subpackage after release evidence.
5. Execute task card 1502 (x/openapi module.yaml cleanup and beta evaluation) once release evidence is available.
6. Clarify `x/data` and `x/fileapi` operational guidance.
7. Expand `x/gateway/discovery` backends only when explicit adapters are ready.

## Roadmap Principles

- Keep `core` as a kernel, not a feature catalog.
- Preserve `net/http` compatibility and explicit control flow.
- Keep one canonical entrypoint per capability family.
- Let `specs/*` and module manifests carry machine-readable authority.
- Document only implemented behavior; remove stale drafts instead of preserving them in the active docs surface.

## What Not to Do

- do not reintroduce component-style compatibility APIs
- do not add new broad legacy roots for short-term convenience
- do not let scenario reference apps replace the canonical app path
- do not move tenant or topology-heavy logic back into stable roots
- do not leave stale historical drafts inside the active docs surface
- do not mark `x/*` packages as GA without explicit policy, tests, and docs
