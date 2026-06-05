# AGENTS.md - plumego

Operational guide for `github.com/spcent/plumego`. Go 1.26+ · stdlib-only main module.

## 1. Authority

Read path: `AGENTS.md` → `specs/task-routing.yaml` entry → `start_with` files → `<module>/module.yaml` → extra docs only when preflight needs.

Conflict order: (1) security/boundary rules here → (2) `docs/reference/canonical-style-guide.md` → (3) machine-readable specs → (4) existing local patterns.

Companion docs: `docs/operations/codex-workflow.md` (workflow), `docs/operations/agent-context-budget.md` (context packages), `docs/operations/agent-code-quality-rules.md` (quality + gate profiles), `docs/reference/canonical-style-guide.md` (style), `docs/concepts/extension-maturity.md` (maturity), `docs/concepts/` (boundary docs).

Machine-readable specs: `task-routing.yaml`, `checks.yaml`, `dependency-rules.yaml`, `extension-taxonomy.yaml`, `module-manifest.schema.yaml`, `stop-condition-handlers.yaml`, `agent-quality-rules.yaml`, `change-recipes/`.

CLI: `go run ./cmd/plumego --help`. Bundles: `make bundle TASK=<recipe> MODULE=<path>`.

## 2. Non-Negotiables

- Preserve `net/http` compatibility.
- Main module stdlib-only unless approved. No `go.mod` under `x/**`.
- `reference/` and `use-cases/` apps may have own `go.mod` with `replace` directive + documented rationale.
- Stable roots must not import `x/*`.
- No hidden globals, `init()` registration, or context service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth/verification/policy errors. Timing-safe compare for secrets.
- Context accessors: `With{Type}` + `{Type}FromContext`; keys are unexported zero-value structs inlined at call site.
- `contract` owns transport primitives only; one canonical success-response and error-construction path per layer.
- Deprecated symbols: remove in same PR as last caller.

## 3. Where To Work

**Stable roots:** `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`

**Extensions:** `x/ai`, `x/data`, `x/fileapi`, `x/frontend`, `x/gateway`, `x/messaging`, `x/observability`, `x/openapi`, `x/resilience`, `x/rest`, `x/rpc`, `x/tenant`, `x/validate`, `x/websocket`

**Landing zones:** kernel/lifecycle/transport/auth/storage → stable root; product capability/protocol/feature → `x/*`; app wiring/DI/bootstrap → `reference/standard-service`; workflow/specs → `docs/` or `specs/`; plans → `tasks/`.

**Maturity:** beta → `frontend`, `gateway`, `messaging`, `observability`, `rest`, `tenant`, `websocket`; experimental → `ai`, `data`, `fileapi`, `openapi`, `resilience`, `rpc`, `validate`. Dashboard: `docs/concepts/extension-maturity.md`, `specs/extension-maturity.yaml`.

**Reference starting points:** `standard-service` (plain JSON), `production-service` (hardened), `with-rest`, `with-tenant`, `with-ai`, `with-websocket`, `with-rpc`, `with-events`, `with-messaging`, `with-webhook`, `with-gateway`, `with-frontend`, `with-ops`, `with-observability`, `with-tenant-admin`, `benchmark`.

**Boundaries:**

- `middleware`: transport-only; no service injection, DTO assembly, or domain-policy branching.
- Tenant/session lifecycle → `x/tenant` (not `middleware`, `store`, `contract`).
- Observability wiring → `x/observability` (not `contract`).
- No new broad top-level roots (`ai`, `tenant`, `net`, `pubsub`, `rest`, `validator`, `utils`, `frontend`).

## 4. Working Contract

Defaults: one primary module, no stable API changes, no new deps, focused tests, docs sync only for behavior/API/config/security/lifecycle/boundary changes.

Use analysis mode (no edits) when: ownership unclear, unstated API change or new dep needed, no acceptance criteria, or spec/manifest/pattern conflict. Resolve via `specs/stop-condition-handlers.yaml`.

Preflight before editing:

```text
Context package:
Owning module:
Target module.yaml read:
In-scope / Out-of-scope paths:
Public API / Dependency / Behavior / Security / Docs impact: none / yes
Validation plan:
```

## 5. Change Rules

- Minimal, scoped to one primary module.
- Read target manifest before editing module behavior.
- Preserve stable public APIs; include migration notes for breaking changes.
- Prefer stdlib + existing local patterns.
- Add/update tests next to changed behavior. No opportunistic refactors.

**Exported symbol changes:** enumerate callers (`rg -n --glob '*.go' 'SymbolName' .`), migrate all in same change, re-verify old name gone, update tests, `go build ./...`. No silent error discard.

## 6. Validation

- **Quick gate:** `make validate-diff` (auto-selects profile by changed paths).
- **Full CI:** `make gates` (vet, format, race tests, coverage, `website/src/generated` staleness, baseline checks below).
- **Docs-only:** skip Go gates unless code/config/generated/examples changed.

Baseline boundary/manifest checks (run manually when needed):

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/cross-extension-deps
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go run ./internal/checks/public-entrypoints-sync
```

Per gate profile (`docs/operations/agent-code-quality-rules.md §6`): add `extension-maturity`, `extension-beta-evidence`, `deprecation-inventory -strict`. Website content/API checks via `make website-gates` (slow).

Summarize: command, status, key failure, next step.

## 7. Docs Sync

Update docs only for behavior, API, config, security, lifecycle, or boundary changes. Targets: `README.md`, `README_CN.md`, `AGENTS.md`, `docs/operations/agent-context-budget.md`, `env.example`, affected `docs/modules/` primers, `docs/concepts/extension-maturity.md`, `docs/release/roadmap.md`, `docs/evidence/stable-api/`.

`docs/modules/` naming: stable roots bare (`core/`); extensions mirror imports (`x/ai/`, `x/data/cache/`). See `docs/modules/INDEX.md`.

Run `make website-sync` after editing sources; include generated files in same commit.

| Source | Generated |
|---|---|
| `docs/release/roadmap.md` | `website/src/generated/roadmap.ts` |
| `docs/modules/*/README.md` | `website/src/generated/modules.ts` |
| `specs/task-routing.yaml` | `website/src/generated/routing.ts` |
| `README.md` / `specs/task-routing.yaml` | `website/src/generated/releases.ts` |
| `website/src/content/docs/**/*.mdx` | `website/src/generated/translation-lag.ts` |

`translation-lag.ts` uses git timestamps; regenerate after every `.mdx` commit.

## 8. Milestones & Cards

- **Milestones:** `tasks/milestones/active/M-NNN-short-name/M-NNN.md`. Execute: read Context first, stay in Affected Modules, follow Tasks in order, use spec branch, record blockers, run acceptance criteria, PR with `docs/assets/github-workflows/milestone-pr-template.md`.
- **Cards:** `tasks/cards/active/NNNN-slug.md`. Focused single-module changes. Active → `tasks/cards/active/`; completed → `tasks/cards/done/`.

Scaffold:

```bash
make new-milestone N=NNN TITLE="..."
make new-plan M=active/M-NNN-...
make new-card ID=NNNN SLUG=... M=M-NNN R=<role>
make new-verify M=active/M-NNN-...
make run-card C=active/NNNN-slug
make milestone-status M=active/M-NNN-...
make check-spec / check-plan / check-verify / check-card ...
```

## 9. Anti-Patterns

Do not introduce:

- New dependencies without approval
- Route auto-discovery or reflection-based wiring
- Middleware that builds business DTOs or injects services
- Ad hoc JSON response helpers or per-feature error envelopes
- New handler signatures
- Panic-only constructors for fallible behavior
- Generic `utils` packages
- Compatibility wrappers without removal plan
