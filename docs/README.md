# Docs Index

This directory is Plumego's human-readable documentation surface.

Read by intent instead of scanning everything.

## Start Here

- `getting-started.md` — smallest runnable example and the canonical next reads
- `getting-started_CN.md` — 中文入门指南（Chinese onboarding guide）
- `ADOPTION_PATH.md` — narrow 5-minute, 30-minute, and 1-day adoption path
- `why-plumego.md` — what Plumego is for, core principles, and fit summary
- `when-not-to-use-plumego.md` — honest cases where a different tool fits better
- `troubleshooting.md` — common problems and fixes (routes, middleware, JWT, lifecycle)
- `benchmarks/README.md` — performance results and methodology vs. Chi, Gin, Echo
- `../reference/standard-service/README.md` — canonical reference app layout and route wiring
- `CANONICAL_STYLE_GUIDE.md` — the default handler, route, middleware, and response style
- `agent-first.md` — agent-first design rationale, mechanisms, and operating reference
- `architecture/AGENT_FIRST_REPO_BLUEPRINT.md` — repository shape, boundaries, and canonical implementation path
- `architecture/core-boundary.md` — stable root package responsibilities and wiring patterns
- `architecture/extension-boundary.md` — x/* taxonomy, maturity ladder, and promotion criteria
- `ROADMAP.md` — current priorities and remaining extension work
- `EXTENSION_MATURITY.md` — current `x/*` status, risk, entrypoint, validation, and promotion blockers
- `stable-api/README.md` — stable-root exported API baseline and regeneration workflow

Canonical onboarding order:

1. `getting-started.md`
2. `why-plumego.md`
3. `ADOPTION_PATH.md`
4. `../reference/standard-service/README.md`
5. this docs index
6. `specs/*` and `tasks/*` when you need machine-readable routing or execution surfaces

## Scenario Entrypoints

Use `../reference/standard-service` as the canonical application shape, then
open the relevant capability primer:

| Scenario | Primary reads |
| --- | --- |
| REST API service | `getting-started.md`, `../reference/standard-service/README.md`, `../reference/with-rest/README.md`, `modules/x-rest/README.md` |
| Multi-tenant API | `../reference/with-tenant/README.md`, `modules/x-tenant/README.md`, `architecture/X_TENANT_BLUEPRINT.md` |
| Edge gateway | `modules/x-gateway/README.md`, including `x/gateway/discovery` when discovery is explicitly selected |
| Realtime service | `modules/x-websocket/README.md`, `modules/x-messaging/README.md` |
| Messaging service | `../reference/with-messaging/README.md`, `modules/x-messaging/README.md`, including `x/messaging/mq` or `x/messaging/pubsub` only for narrow primitive work |
| Webhook ingress or delivery | `../reference/with-webhook/README.md`, `modules/x-messaging/README.md`, including `x/messaging/webhook` for narrow transport work |
| RPC transport | `../reference/with-rpc/README.md`, `modules/x-rpc/README.md`, starting with `x/rpc/server`, `x/rpc/client`, or `x/rpc/gateway` for the direct transport task |
| File API service | `../reference/standard-service/README.md`, `modules/x-fileapi/README.md`, `modules/x-data/README.md` for storage and metadata backends |
| AI service | `../reference/with-ai/README.md`, `modules/x-ai/README.md`, starting with provider, session, streaming, and tool subpackages |
| Observability and ops | `../reference/with-ops/README.md`, `modules/x-observability/README.md`, `modules/x-ops/README.md`, `modules/x-devtools/README.md` |
| Cross-family resilience primitives | `modules/x-resilience/README.md`, and `architecture/x-ai-resilience-boundary.md` when AI wrappers are involved |

These paths identify first reads. They do not replace the canonical bootstrap
layout or promote experimental `x/*` APIs to stable status.

## Architecture

- `architecture/AGENT_FIRST_REPO_BLUEPRINT.md` — overall repository shape and control-plane split
- `architecture/core-boundary.md` — per-package responsibility table, wiring patterns, and decision checklist
- `architecture/extension-boundary.md` — x/* taxonomy, four-category classification, maturity ladder, promotion criteria
- `architecture/X_TENANT_BLUEPRINT.md` — tenant architecture and isolation model

## Workflow

- `CODEX_WORKFLOW.md` — milestone execution guide tied to the current `Makefile`
- `AGENT_CODE_QUALITY_RULES.md` — agent preflight, review output, risk focus, and gate selection rules
- `MILESTONE_PIPELINE.md` — artifact contract for milestone, plan, card, verify, and PR handoffs
- `DEPRECATION.md` — compatibility and deprecation policy
- `EXTENSION_MATURITY.md` — extension family maturity dashboard
- `SCAFFOLD_REFERENCE_CONTRACT.md` — canonical scaffold and reference app alignment contract
- `release/PRE_V1_RELEASE_CHECKLIST.md` — release gate evidence checklist

Related execution surfaces live outside `docs/`:

- `tasks/milestones/README.md` — milestone queue usage and lifecycle
- `tasks/cards/README.md` — task-card queue usage and ownership rules
- `specs/*` — machine-readable routing, taxonomy, dependency, quality, and validation rules

## Module Primers

- `modules/*/README.md` — module-family primers that should stay aligned with each manifest's `doc_paths`
- Stable-root primers cover `contract`, `core`, `router`, `middleware`,
  `security`, `store`, `health`, `log`, and `metrics`.
- Extension primers cover the current `x/*` families declared in `specs/repo.yaml`.
  Start from the primary family primer when a subordinate package exists
  (`x/messaging` before `x/messaging/mq` or `x/messaging/pubsub`, `x/data` before `x/data/cache`,
  `x/observability` before `x/observability/ops` or `x/observability/devtools`).
- Start from `modules/x-resilience/README.md` when the task is reusable
  circuit-breaker or rate-limit behavior shared across extension families.

## Reference Apps

- `../reference/standard-service/README.md` — canonical application layout and the default first read
- `../reference/production-service/README.md` — production-oriented explicit wiring beyond the minimal canonical path
- `../reference/with-rest/README.md` — explicit `x/rest` mounting pattern
- `../reference/with-tenant/README.md` — explicit `x/tenant` mounting pattern
- `../reference/with-gateway/README.md` — explicit `x/gateway` mounting pattern
- `../reference/with-websocket/README.md` — explicit `x/websocket` mounting pattern
- `../reference/with-ai/README.md` — explicit stable-tier `x/ai` mounting pattern
- `../reference/with-ops/README.md` — explicit observability and protected admin surface mounting pattern
- `../reference/with-messaging/README.md` — explicit app-facing messaging mounting pattern
- `../reference/with-webhook/README.md` — explicit webhook ingress and outbound delivery mounting pattern
- `../reference/workerfleet/README.md` — separate worker-oriented reference, not part of the canonical web-service path

## Reference Assets

- `github-workflows/milestone-pr-template.md` — PR body template for milestone work
- `github-workflows/milestone-gates.yml` — example milestone-only workflow for downstream or copied setups

The live repository CI workflow is `.github/workflows/quality-gates.yml`.

## Release And Evidence

- `EXTENSION_STABILITY_POLICY.md` — promotion criteria for `experimental` → `beta` → `ga`
- `EXTENSION_MATURITY.md` — current extension status and blocker dashboard
- `extension-evidence/` — per-family evidence ledgers and snapshot-backed promotion notes
- `stable-api/README.md` — stable-root exported API baseline workflow
- `release/PRE_V1_RELEASE_CHECKLIST.md` — release gate checklist
- `release/v1.0.0-rc.1.md` and `release/v1.0.0.md` — release notes and evidence summaries

## Authority Order

When guidance overlaps, use this order:

1. `AGENTS.md`
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. `docs/AGENT_CODE_QUALITY_RULES.md`
4. `specs/*`
5. the touched module's `module.yaml`
6. existing local patterns in touched files
