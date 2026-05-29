# AGENTS.md - plumego

Operational guide for AI coding agents working in `github.com/spcent/plumego`.

Go module: `github.com/spcent/plumego` | Go 1.26+ | Main module: stdlib only unless approved

## 1. Authority

Default read path: `AGENTS.md` → matching `specs/task-routing.yaml` entry → its `start_with` files → target `<module>/module.yaml` → extra docs only when preflight identifies a concrete need.

Companion docs: `docs/CODEX_WORKFLOW.md` (workflow), `docs/AGENT_CONTEXT_BUDGET.md` (context packages), `docs/AGENT_CODE_QUALITY_RULES.md` (quality + gate profiles), `docs/CANONICAL_STYLE_GUIDE.md` (style), `docs/EXTENSION_MATURITY.md` (maturity), `docs/architecture/` (boundary docs).

Machine-readable: `specs/task-routing.yaml`, `specs/checks.yaml`, `specs/dependency-rules.yaml`, `specs/extension-taxonomy.yaml`, `specs/module-manifest.schema.yaml`, `specs/stop-condition-handlers.yaml`, `specs/agent-quality-rules.yaml`, `specs/change-recipes/`. Canonical wiring: `reference/standard-service`.

CLI tool: `cmd/plumego` provides agent-assist commands, validation runners, code generation, dev server, scaffold, and task bundling. Run `go run ./cmd/plumego --help` to explore subcommands; `make bundle TASK=<recipe> MODULE=<path>` generates task execution bundles.

Conflict order: (1) security/boundary rules here → (2) `docs/CANONICAL_STYLE_GUIDE.md` → (3) machine-readable specs → (4) existing local patterns.

## 2. Non-Negotiables

- Preserve `net/http` compatibility.
- Keep the main module dependency-free beyond the standard library unless explicitly approved.
- Do not add `go.mod` anywhere under `x/**`; extension packages remain part of the main module.
- Reference apps under `reference/` MAY have an independent `go.mod` for external deps; use a `replace` directive and document the rationale. Use-case apps under `use-cases/` follow the same rule (e.g. `use-cases/workerfleet` for MongoDB).
- Stable roots must not import `x/*`.
- Do not introduce hidden globals, `init()` registration, or context service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth, verification, and policy errors.
- Use timing-safe comparison for secret checks.
- Context accessors use `With{Type}` plus `{Type}FromContext`; key types are unexported zero-value structs inlined at the call site.
- `contract` owns transport primitives only (response/error helpers, metadata, accessors, binding); one canonical success-response and error-construction path per layer.
- Deprecated symbols must be removed in the same PR that replaces their last caller. Do not leave dead wrappers behind.

## 3. Where To Work

Stable roots: `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`

Extension roots: `x/ai`, `x/data`, `x/fileapi`, `x/frontend`, `x/gateway`, `x/messaging`, `x/observability`, `x/openapi`, `x/resilience`, `x/rest`, `x/rpc`, `x/tenant`, `x/validate`, `x/websocket`

Default landing zones: kernel/lifecycle/transport/auth/storage → stable root; product capability/protocol/features → `x/*`; app wiring/DI/bootstrap → `reference/standard-service`; workflow/specs/quality → `docs/` or `specs/`; plans/sequencing → `tasks/`.

Extension maturity: **beta** (production-ready with caveats) → `x/frontend`, `x/gateway`, `x/messaging`, `x/observability`, `x/rest`, `x/tenant`, `x/websocket`; **experimental** (APIs may change) → `x/ai`, `x/data`, `x/fileapi`, `x/openapi`, `x/resilience`, `x/rpc`, `x/validate`. Full dashboard: `docs/EXTENSION_MATURITY.md` and `specs/extension-maturity.yaml`.

Reference starting points: plain JSON API → `reference/standard-service`; hardened production (auth, tracing, metrics, tenant) → `reference/production-service`; REST CRUD → `reference/with-rest`; multi-tenant → `reference/with-tenant`; LLM/AI → `reference/with-ai`; WebSocket → `reference/with-websocket`; gRPC → `reference/with-rpc`; event-driven pubsub architecture → `reference/with-events`; messaging feature integration into existing service → `reference/with-messaging`; inbound webhooks → `reference/with-webhook`; reverse proxy → `reference/with-gateway`; embedded static assets → `reference/with-frontend`; health/metrics routes → `reference/with-ops`; benchmarking → `reference/benchmark`.

Boundary reminders:

- `middleware` stays transport-only; no service injection, business DTO assembly, or domain-policy branching.
- Tenant logic and session lifecycle belong in `x/tenant`; not in `middleware`, `store`, or `contract`.
- Observability wiring (tracing, metrics export) belongs in `x/observability`, not `contract`.
- New library code must live under a stable root or `x/*`; avoid new broad top-level roots such as `ai`, `tenant`, `net`, `pubsub`, `rest`, `validator`, `utils`, or `frontend`.

## 4. Working Contract

Default assumptions: one primary module, no stable API changes, no new dependencies, focused tests, docs sync only for behavior/API/config/security/lifecycle/boundary changes.

Use analysis mode (no edits) when ownership is unclear, an unstated API change or new dependency is needed, the task lacks acceptance criteria, or a spec/manifest/pattern conflict exists. Use `specs/stop-condition-handlers.yaml` for stop-condition resolution.

Before editing, complete this preflight:

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

- Keep changes minimal and scoped to one primary module when possible.
- Read the target manifest before editing module behavior.
- Preserve stable public APIs unless explicitly asked to change them.
- If a breaking change is unavoidable, include migration notes.
- Prefer standard-library solutions and existing local patterns.
- Add or update tests next to changed behavior.
- Do not refactor unrelated files opportunistically.

Exported symbol changes: enumerate callers (`rg -n --glob '*.go' 'SymbolName' .`), migrate all in the same change, re-run search to confirm old name is gone, update tests, `go build ./...`. No caller may silently discard a newly returning error.

## 6. Validation

Quick gate for current diff: `make validate-diff` (auto-selects minimal profile based on changed paths). Full CI: `make gates`.

Run: (1) target module checks from `<module>/module.yaml`, (2) boundary checks, (3) repo-wide gates only when gate profile requires.

Baseline boundary and manifest checks:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/cross-extension-deps
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go run ./internal/checks/public-entrypoints-sync
```

Add `extension-maturity`, `extension-beta-evidence`, `deprecation-inventory -strict` per gate profile (`docs/AGENT_CODE_QUALITY_RULES.md §6`). Full CI: `make gates` (runs above + `go vet ./...`, format, race tests; run `gofmt -w` first). Docs-only: skip Go gates unless code/config/generated/examples changed.

Summarize validation as: command, status, key failure, next step.

## 7. Docs Sync

Update docs only for behavior, API, config, security, lifecycle, or boundary changes. Typical targets: `README.md`, `README_CN.md`, `AGENTS.md`, `docs/AGENT_CONTEXT_BUDGET.md`, `env.example`, affected `docs/modules/` primers, `docs/EXTENSION_MATURITY.md`, `docs/ROADMAP.md`, `docs/stable-api/`. Document implemented behavior only.

`docs/modules/` naming: stable roots use bare names (`core/`, `contract/`); extensions mirror import paths (`x/ai/`, `x/data/cache/`). See `docs/modules/INDEX.md`. Run `make website-sync` after editing sources below; include generated files in the same commit.

| Source file | Generated file |
|---|---|
| `docs/ROADMAP.md` | `website/src/generated/roadmap.ts` |
| `docs/modules/*/README.md` | `website/src/generated/modules.ts` |
| `specs/task-routing.yaml` | `website/src/generated/routing.ts` |
| `README.md` / `specs/task-routing.yaml` | `website/src/generated/releases.ts` |
| any `website/src/content/docs/**/*.mdx` (en or zh) | `website/src/generated/translation-lag.ts` |

`translation-lag.ts` is from git timestamps; regenerate after every `.mdx` commit.

## 8. Milestones

Milestones: `tasks/milestones/active/M-NNN-short-name/M-NNN.md`. When executing: read Context files first, stay inside Affected Modules, follow Tasks in order, use spec branch, record blockers, run full acceptance criteria, package PR with `docs/github-workflows/milestone-pr-template.md`.

Scaffold: `make new-milestone`, `make new-plan`, `make new-card`, `make new-verify`, `make run-card C=active/NNNN-slug`, `make milestone-status M=active/M-NNN`.

Task cards: `tasks/cards/active/NNNN-slug.md`. Cards are narrower than milestones — use them for focused, time-boxed changes within a single module. Execute: `make run-card C=active/NNNN-slug` (validates, bundles, and runs via codex). Active cards live in `tasks/cards/active/`; completed in `tasks/cards/done/`.

## 9. Anti-Patterns

Do not introduce:

- New dependencies without approval
- Route auto-discovery or reflection-based wiring
- Middleware that builds business DTOs or injects services
- Ad hoc JSON response helpers or per-feature error envelopes
- New handler signatures
- New panic-only constructors for fallible behavior
- Generic `utils` packages
- Compatibility wrappers without a removal plan
