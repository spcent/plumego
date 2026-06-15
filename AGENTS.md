# AGENTS.md - plumego

Operational guide for AI coding agents in `github.com/spcent/plumego`. Go 1.26+; main module stdlib-only unless approved.

## 1. Authority

Read path: `AGENTS.md` → matching `specs/task-routing.yaml` entry → its `start_with` files → target `<module>/module.yaml` → extra docs only when preflight identifies a concrete need. Conflict order: (1) security/boundary rules here → (2) `docs/reference/canonical-style-guide.md` → (3) machine-readable specs → (4) existing local patterns.

Docs: `docs/operations/` (gate profiles, workflow, context packages), `docs/reference/canonical-style-guide.md`, `docs/concepts/` (boundaries, extension maturity). Specs: `specs/` (`task-routing.yaml`, `dependency-rules.yaml`, `stop-condition-handlers.yaml`, `change-recipes/`). Canonical wiring: `reference/standard-service`.

CLI: `go run ./cmd/plumego --help` (agent-assist, validation runners, codegen, dev server, scaffold, task bundling); `make bundle TASK=<recipe> MODULE=<path>`.

## 2. Non-Negotiables

- Preserve `net/http` compatibility.
- Main module stdlib-only unless approved; no `go.mod` under `x/**`; stable roots must not import `x/*`.
- `reference/` and `use-cases/` MAY have an independent `go.mod` for external deps via a `replace` directive + documented rationale (`workerfleet`=MongoDB, `cloud-vault`=multi-feature wiring, `dbadmin`=admin tooling, `guardus`=security services, `mini-saas-api`=stable-roots + beta-extension SaaS showcase with zero external deps).
- No hidden globals, `init()` registration, or context service-locator patterns.
- Never log secrets/tokens/signatures/keys. Fail closed on auth/verification/policy errors; timing-safe comparison for secret checks.
- Context accessors: `With{Type}` + `{Type}FromContext`; key types are unexported zero-value structs inlined at the call site.
- `contract` owns transport primitives only (response/error helpers, metadata, accessors, binding); one canonical success-response and error path per layer.
- Remove deprecated symbols in the same PR that replaces their last caller; leave no dead wrappers.

## 3. Where To Work

Stable roots: `core router contract middleware security store health log metrics`.

Extensions (`x/`): **beta** (production-ready, caveats) = `frontend gateway messaging observability rest tenant websocket`; **experimental** (APIs may change) = `ai data fileapi openapi resilience rpc validate`. Dashboard: `docs/concepts/extension-maturity.md`, `specs/extension-maturity.yaml`.

Landing zones: kernel/lifecycle/transport/auth/storage → stable root; product capability/protocol/feature → `x/*`; app wiring/DI/bootstrap → `reference/standard-service`; workflow/specs/quality → `docs/` or `specs/`; plans/sequencing → `tasks/`. New library code lives under a stable root or `x/*`; no new broad top-level roots (`ai`, `tenant`, `net`, `pubsub`, `rest`, `validator`, `utils`, `frontend`).

Reference starts (`reference/`): JSON API→`standard-service`; hardened prod→`production-service`; REST CRUD→`with-rest`; multi-tenant→`with-tenant`; LLM/AI→`with-ai`; WebSocket→`with-websocket`; gRPC→`with-rpc`; observability→`with-observability`. Full map: `specs/task-routing.yaml`.

Boundaries: `middleware` is transport-only (no service injection, DTO assembly, or domain-policy branching); tenant logic + session lifecycle → `x/tenant` (not middleware/store/contract); observability wiring (tracing, metrics export) → `x/observability` (not contract).

## 4. Working Contract

Defaults: one primary module, no stable-API change, no new deps, focused tests, docs sync only for behavior/API/config/security/lifecycle/boundary changes.

Analysis mode (no edits) when: ownership unclear, an unstated API change or new dependency is needed, the task lacks acceptance criteria, or a spec/manifest/pattern conflict exists. Resolve stops via `specs/stop-condition-handlers.yaml`.

Preflight before editing — state each: Context package; Owning module; Target module.yaml read; In-scope paths; Out-of-scope paths; Public API / Dependency / Behavior / Security / Docs impact (none|yes); Validation plan.

## 5. Change Rules

- Minimal changes, scoped to one primary module; read its manifest before editing behavior.
- Preserve stable public APIs unless explicitly asked; migration notes for unavoidable breaks. No caller may silently discard a newly returning error.
- Prefer stdlib + existing local patterns; add/update tests next to changed behavior. No opportunistic refactors of unrelated files.
- Exported symbol changes: enumerate callers (`rg -n --glob '*.go' 'SymbolName' .`), migrate all in the same change, re-search to confirm the old name is gone, update tests, `go build ./...`.

## 6. Validation

Quick diff gate: `make validate-diff` (minimal profile). Full CI: `make gates`. Run: (1) target module checks from `<module>/module.yaml`, (2) boundary checks, (3) repo-wide gates only when the gate profile requires.

Baseline boundary/manifest checks (`go run ./internal/checks/<name>`): `dependency-rules cross-extension-deps agent-workflow module-manifests reference-layout public-entrypoints-sync`. Add `extension-maturity`, `extension-beta-evidence`, `deprecation-inventory -strict` per gate profile (`docs/operations/agent-code-quality-rules.md §6`).

`make gates` = above + `go vet ./...`, format, race tests, coverage, `website/src/generated` staleness check (run `gofmt -w` first). `make website-gates` (slow) = website content/API checks + static build. Docs-only: skip Go gates unless code/config/generated/examples changed. Summarize validation as: command, status, key failure, next step.

## 7. Docs Sync

Update docs only for behavior/API/config/security/lifecycle/boundary changes; document implemented behavior only. Targets: `README.md`, `README_CN.md`, `AGENTS.md`, `docs/operations/agent-context-budget.md`, `env.example`, affected `docs/modules/` primers (see `docs/modules/INDEX.md` for naming), `docs/concepts/extension-maturity.md`, `docs/release/roadmap.md`, `docs/evidence/stable-api/`.

Run `make website-sync` after editing these sources, then commit the generated files (all under `website/src/generated/`) together:

| Source | Generated |
|---|---|
| `docs/release/roadmap.md` | `roadmap.ts` |
| `docs/modules/*/README.md` | `modules.ts` |
| `specs/task-routing.yaml` | `routing.ts` |
| `README.md` + `specs/task-routing.yaml` | `releases.ts` |
| `website/src/content/docs/**/*.mdx` (en/zh) | `translation-lag.ts` |

`translation-lag.ts` derives from git timestamps; regenerate after every `.mdx` commit.

## 8. Milestones

Milestones: `tasks/milestones/active/M-NNN-short-name/M-NNN.md` — read Context files first, stay inside Affected Modules, follow Tasks in order, use the spec branch, record blockers, run full acceptance criteria, package the PR via `docs/assets/github-workflows/milestone-pr-template.md`. Task cards: `tasks/cards/active/NNNN-slug.md` (narrower, single-module; → `done/` when complete). Scaffold: `make new-milestone|new-plan|new-card|new-verify`, `run-card C=active/NNNN-slug`, `milestone-status M=active/M-NNN`.

## 9. Anti-Patterns

No: new deps without approval; route auto-discovery / reflection-based wiring; middleware that builds DTOs or injects services; ad hoc JSON response helpers or per-feature error envelopes; new handler signatures; new panic-only constructors for fallible behavior; generic `utils` packages; compatibility wrappers without a removal plan.
