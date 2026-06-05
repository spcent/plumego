# AGENTS.md - plumego

Operational guide for AI coding agents in `github.com/spcent/plumego`. Go 1.26+. Main module: stdlib only unless approved.

## 1. Authority

Read path: `AGENTS.md` → matching `specs/task-routing.yaml` entry → its `start_with` files → target `<module>/module.yaml` → extra docs only when preflight identifies a concrete need.

Companion docs: `docs/operations/codex-workflow.md` (workflow), `docs/operations/agent-context-budget.md` (context packages), `docs/operations/agent-code-quality-rules.md` (quality + gate profiles), `docs/reference/canonical-style-guide.md` (style), `docs/concepts/extension-maturity.md` (maturity), `docs/concepts/` (boundaries).

Machine-readable specs under `specs/`: `task-routing.yaml`, `checks.yaml`, `dependency-rules.yaml`, `extension-taxonomy.yaml`, `module-manifest.schema.yaml`, `stop-condition-handlers.yaml`, `specs/agent-quality-rules.yaml`, `change-recipes/`. Canonical wiring: `reference/standard-service`.

CLI `cmd/plumego`: agent-assist, validation runners, codegen, dev server, scaffold, task bundling. `go run ./cmd/plumego --help`; `make bundle TASK=<recipe> MODULE=<path>` builds task bundles.

Conflict order: (1) security/boundary rules here → (2) `canonical-style-guide.md` → (3) machine-readable specs → (4) existing local patterns.

## 2. Non-Negotiables

- Preserve `net/http` compatibility.
- Main module stays dependency-free beyond stdlib unless explicitly approved.
- No `go.mod` under `x/**`; extensions stay part of the main module.
- Reference apps (`reference/`) and use-case apps (`use-cases/`) MAY have an independent `go.mod` for external deps via a `replace` directive with documented rationale (e.g. `workerfleet`=MongoDB, `cloud-vault`=multi-feature wiring, `dbadmin`=admin tooling, `guardus`=security services).
- Stable roots must not import `x/*`.
- No hidden globals, `init()` registration, or context service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth, verification, and policy errors. Use timing-safe comparison for secret checks.
- Context accessors: `With{Type}` + `{Type}FromContext`; key types are unexported zero-value structs inlined at the call site.
- `contract` owns transport primitives only (response/error helpers, metadata, accessors, binding); one canonical success-response and error path per layer.
- Remove deprecated symbols in the same PR that replaces their last caller; leave no dead wrappers.

## 3. Where To Work

Stable roots: `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`.

Extension roots (`x/`): `ai`, `data`, `fileapi`, `frontend`, `gateway`, `messaging`, `observability`, `openapi`, `resilience`, `rest`, `rpc`, `tenant`, `validate`, `websocket`.

Landing zones: kernel/lifecycle/transport/auth/storage → stable root; product capability/protocol/features → `x/*`; app wiring/DI/bootstrap → `reference/standard-service`; workflow/specs/quality → `docs/` or `specs/`; plans/sequencing → `tasks/`.

Maturity: **beta** (production-ready, caveats) = `frontend`, `gateway`, `messaging`, `observability`, `rest`, `tenant`, `websocket`; **experimental** (APIs may change) = `ai`, `data`, `fileapi`, `openapi`, `resilience`, `rpc`, `validate`. Dashboard: `docs/concepts/extension-maturity.md`, `specs/extension-maturity.yaml`.

Reference starting points: JSON API → `standard-service`; hardened prod (auth/tracing/metrics/tenant) → `production-service`; REST CRUD → `with-rest`; multi-tenant → `with-tenant`; LLM/AI → `with-ai`; WebSocket → `with-websocket`; gRPC → `with-rpc`; event/pubsub → `with-events`; messaging integration → `with-messaging`; webhooks → `with-webhook`; reverse proxy → `with-gateway`; embedded assets → `with-frontend`; health/metrics → `with-ops`; observability stack → `with-observability`; tenant admin console → `with-tenant-admin`; benchmarking → `benchmark` (all under `reference/`).

Boundaries:

- `middleware` is transport-only; no service injection, DTO assembly, or domain-policy branching.
- Tenant logic and session lifecycle live in `x/tenant`, not `middleware`/`store`/`contract`.
- Observability wiring (tracing, metrics export) lives in `x/observability`, not `contract`.
- New library code lives under a stable root or `x/*`; no new broad top-level roots (`ai`, `tenant`, `net`, `pubsub`, `rest`, `validator`, `utils`, `frontend`).

## 4. Working Contract

Default assumptions: one primary module, no stable API changes, no new deps, focused tests, docs sync only for behavior/API/config/security/lifecycle/boundary changes.

Use analysis mode (no edits) when: ownership is unclear, an unstated API change or new dependency is needed, the task lacks acceptance criteria, or a spec/manifest/pattern conflict exists. Resolve stops via `specs/stop-condition-handlers.yaml`.

Preflight before editing:

```text
Context package:
Owning module:
Target module.yaml read:
In-scope paths:
Out-of-scope paths:
Public API impact: none / yes
Dependency impact: none / yes
Behavior impact: none / yes
Security impact: none / yes
Docs impact: none / yes
Validation plan:
```

## 5. Change Rules

- Keep changes minimal, scoped to one primary module when possible.
- Read the target manifest before editing module behavior.
- Preserve stable public APIs unless explicitly asked to change them; include migration notes for unavoidable breaks.
- Prefer stdlib solutions and existing local patterns. Add/update tests next to changed behavior.
- Do not refactor unrelated files opportunistically.
- Exported symbol changes: enumerate callers (`rg -n --glob '*.go' 'SymbolName' .`), migrate all in the same change, re-search to confirm the old name is gone, update tests, `go build ./...`. No caller may silently discard a newly returning error.

## 6. Validation

Quick diff gate: `make validate-diff` (auto-selects minimal profile). Full CI: `make gates`.

Run: (1) target module checks from `<module>/module.yaml`, (2) boundary checks, (3) repo-wide gates only when the gate profile requires.

Baseline boundary/manifest checks:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/cross-extension-deps
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go run ./internal/checks/public-entrypoints-sync
```

Add `extension-maturity`, `extension-beta-evidence`, `deprecation-inventory -strict` per gate profile (`docs/operations/agent-code-quality-rules.md §6`). `make gates` = above + `go vet ./...`, format, race tests, coverage, and `website/src/generated` staleness check (run `gofmt -w` first). Website content/API checks + static build run separately via `make website-gates` (slow). Docs-only: skip Go gates unless code/config/generated/examples changed.

Summarize validation as: command, status, key failure, next step.

## 7. Docs Sync

Update docs only for behavior/API/config/security/lifecycle/boundary changes; document implemented behavior only. Typical targets: `README.md`, `README_CN.md`, `AGENTS.md`, `docs/operations/agent-context-budget.md`, `env.example`, affected `docs/modules/` primers, `docs/concepts/extension-maturity.md`, `docs/release/roadmap.md`, `docs/evidence/stable-api/`.

`docs/modules/` naming: stable roots bare (`core/`); extensions mirror import paths (`x/ai/`, `x/data/cache/`). See `docs/modules/INDEX.md`. Run `make website-sync` after editing the sources below; commit generated files together.

| Source | Generated |
|---|---|
| `docs/release/roadmap.md` | `website/src/generated/roadmap.ts` |
| `docs/modules/*/README.md` | `website/src/generated/modules.ts` |
| `specs/task-routing.yaml` | `website/src/generated/routing.ts` |
| `README.md` / `specs/task-routing.yaml` | `website/src/generated/releases.ts` |
| `website/src/content/docs/**/*.mdx` (en/zh) | `website/src/generated/translation-lag.ts` |

`translation-lag.ts` derives from git timestamps; regenerate after every `.mdx` commit.

## 8. Milestones

Milestones: `tasks/milestones/active/M-NNN-short-name/M-NNN.md`. Executing: read Context files first, stay inside Affected Modules, follow Tasks in order, use the spec branch, record blockers, run full acceptance criteria, package the PR with `docs/assets/github-workflows/milestone-pr-template.md`.

Task cards: `tasks/cards/active/NNNN-slug.md` — narrower than milestones, for focused single-module changes (`done/` when complete).

Scaffold: `make new-milestone`, `new-plan`, `new-card`, `new-verify`, `run-card C=active/NNNN-slug` (validates, bundles, runs via codex), `milestone-status M=active/M-NNN`.

## 9. Anti-Patterns

Do not introduce: new deps without approval; route auto-discovery or reflection-based wiring; middleware that builds DTOs or injects services; ad hoc JSON response helpers or per-feature error envelopes; new handler signatures; new panic-only constructors for fallible behavior; generic `utils` packages; compatibility wrappers without a removal plan.
