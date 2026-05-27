# AGENTS.md - plumego

Operational guide for AI coding agents working in `github.com/spcent/plumego`.

Plumego is an agent-first Go toolkit built on the standard-library HTTP model.
Keep the stable surface small, ownership obvious, and every change easy to
review and reverse.

Go module: `github.com/spcent/plumego`
Go version: `go 1.26.0` with toolchain `go1.26.0` or a newer patch release
Main module policy: standard library only unless explicitly approved

## 1. Authority

Use the smallest safe context package for the task.

Default read path:

1. Read `AGENTS.md`.
2. Select the matching `specs/task-routing.yaml` task entry.
3. Read only that entry's `start_with` files.
4. Read the target `<module>/module.yaml` before editing module behavior.
5. Read extra docs or specs only when preflight identifies a concrete need.

Control-plane companions:

- Workflow: `docs/CODEX_WORKFLOW.md`
- Context budget: `docs/AGENT_CONTEXT_BUDGET.md`
- Quality contract: `docs/AGENT_CODE_QUALITY_RULES.md`
- Style: `docs/CANONICAL_STYLE_GUIDE.md`
- Architecture: `docs/agent-first-operating-reference.md`, `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`, `docs/architecture/core-boundary.md`, `docs/architecture/extension-boundary.md`
- Maturity: `docs/EXTENSION_MATURITY.md`
- Machine-readable rules: `specs/repo.yaml`, `specs/task-routing.yaml`, `specs/extension-taxonomy.yaml`, `specs/extension-maturity.yaml`, `specs/package-hotspots.yaml`, `specs/dependency-rules.yaml`, `specs/checks.yaml`, `specs/module-manifest.schema.yaml`, `specs/stop-condition-handlers.yaml`, `specs/request-flows.yaml`, `specs/agent-quality-rules.yaml`, `specs/change-recipes/`
- Canonical app wiring: `reference/standard-service`

When guidance conflicts, follow this order:

1. Security and boundary rules in this file
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. Machine-readable specs
4. Existing local patterns in touched files

## 2. Non-Negotiables

- Preserve `net/http` compatibility.
- Keep the main module dependency-free beyond the standard library unless explicitly approved.
- Do not add `go.mod` anywhere under `x/**`; extension packages remain part of the main module.
- Reference apps under `reference/` MAY have an independent `go.mod` only when they require external dependencies (e.g. MongoDB, driver-specific packages) that must not enter the root module. The one current example is `reference/workerfleet`, which uses MongoDB and references the root module via a `replace` directive. Any new reference app requiring this exception must document the rationale in its own README and use the same `replace` pattern.
- Stable roots must not import `x/*`.
- Do not introduce hidden globals, `init()` registration, or context service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth, verification, and policy errors.
- Use timing-safe comparison for secret checks.
- Context accessors use `With{Type}` plus `{Type}FromContext`; key types are unexported zero-value structs inlined at the call site.
- `contract` owns transport primitives only: response and error helpers, request metadata, context accessors, and binding helpers.
- Keep one canonical success-response path and one canonical error-construction path per layer.
- Deprecated symbols must be removed in the same PR that replaces their last caller. Do not leave dead wrappers behind.

## 3. Where To Work

Stable roots:

- `core`
- `router`
- `contract`
- `middleware`
- `security`
- `store`
- `health`
- `log`
- `metrics`

Extension roots:

- `x/ai`
- `x/data`
- `x/fileapi`
- `x/frontend`
- `x/gateway`
- `x/messaging`
- `x/observability`
- `x/openapi`
- `x/resilience`
- `x/rest`
- `x/rpc`
- `x/tenant`
- `x/validate`
- `x/websocket`

Default landing zones:

- Kernel, lifecycle, routing, transport contracts, transport middleware, auth primitives, and storage primitives: stable root
- Product capability, protocol adaptation, and feature behavior: `x/*`
- Application wiring, bootstrap, DI, and route registration: `reference/standard-service` or `cmd/plumego/internal/scaffold`
- Workflow rules, boundaries, quality gates, and specs: `docs/` or `specs/`
- Execution plans and task sequencing: `tasks/cards/` or `tasks/milestones/`

Boundary reminders:

- `core` is the app kernel, not a feature catalog.
- `router` owns matching, params, groups, and reverse routing.
- `middleware` stays transport-only; no service injection, business DTO assembly, or domain-policy branching.
- `health` owns health models and readiness helpers, not HTTP handler ownership.
- Tenant logic belongs in `x/tenant`, not stable `middleware` or `store`.
- Tracing, sampling, collectors, and metric export wiring belong in `x/observability`, not `contract`.
- Session lifecycle belongs in `x/tenant` or an owning extension, not `contract`.
- New library code must live under a stable root or `x/*`; avoid new broad top-level roots such as `ai`, `tenant`, `net`, `pubsub`, `rest`, `validator`, `utils`, or `frontend`.

## 4. Working Contract

Default assumptions for normal implementation work:

- One primary module per change
- No stable public API changes
- No new dependencies
- Focused tests for behavior changes
- Docs sync only for behavior, API, config, security, lifecycle, or boundary changes

Use analysis mode and do not edit when:

- Ownership is unclear
- A stable public API change was not requested
- A new dependency is required
- The task is broad without acceptance criteria
- A repo spec, module manifest, and local pattern conflict in a behavior-changing way

When you hit a stop condition, use `specs/stop-condition-handlers.yaml`.

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

Exported symbol changes must be complete in one change:

1. Enumerate callers with `rg -n --glob '*.go' 'SymbolName' .`.
2. Migrate or update every caller in the same change.
3. Re-run the search; the old name must be gone for deletions and renames.
4. Update tests in the same commit.
5. Run `go build ./...` and the required tests before handoff.

No caller may silently discard a newly returning function without an explicit discard.

## 6. Validation

Validation order:

1. Run target module checks from `<module>/module.yaml`.
2. Run boundary and manifest checks.
3. Run repo-wide gates only when the selected gate profile requires them.

Baseline boundary and manifest checks:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/cross-extension-deps
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go run ./internal/checks/public-entrypoints-sync
```

Add `extension-maturity`, `extension-beta-evidence`, and `deprecation-inventory -strict`
when the gate profile in `docs/AGENT_CODE_QUALITY_RULES.md` or
`specs/agent-quality-rules.yaml` requires them.

Repo-wide gate:

```bash
make gates
```

`make gates` mirrors CI: boundary checks, `go vet ./...`, non-mutating
formatting checks, race tests, and normal tests. Run `gofmt -w <paths>` first
when needed.

Docs-only changes: check accuracy, links, terminology, and authority order. Go
gates are not required unless code, config, generated data, or runnable
examples changed.

Summarize validation with command, status, key failure, and next step instead
of pasting full logs.

## 7. Docs Sync

Update docs only when behavior, public API, config, security semantics,
lifecycle behavior, or boundaries change. Common sync targets:

- `README.md`
- `README_CN.md`
- `AGENTS.md`
- `docs/AGENT_CONTEXT_BUDGET.md`
- `env.example`

Also sync affected module primers under `docs/modules/`, `docs/EXTENSION_MATURITY.md`,
`docs/ROADMAP.md`, or `docs/stable-api/` when that surface changes.

The `docs/modules/` naming convention: stable roots use bare names (`core/`, `contract/`);
extension packages mirror the actual import path under a top-level `x/` directory
(`x/ai/`, `x/gateway/`, `x/data/cache/`, `x/messaging/mq/`). The directory structure
in `docs/modules/x/` is a 1:1 reflection of the `x/` source tree. See
`docs/modules/INDEX.md` for the full mapping table.

Document implemented behavior only.

### Website Generated Files

`website/src/generated/` is checked in and must stay in sync with its sources.
After editing any source below, run `make website-sync` and include the updated
generated files in the same commit.

| Source file | Generated file |
|---|---|
| `docs/ROADMAP.md` | `website/src/generated/roadmap.ts` |
| `docs/modules/*/README.md` | `website/src/generated/modules.ts` |
| `specs/task-routing.yaml` | `website/src/generated/routing.ts` |
| `README.md` / `specs/task-routing.yaml` | `website/src/generated/releases.ts` |
| any `website/src/content/docs/**/*.mdx` (en or zh) | `website/src/generated/translation-lag.ts` |

`translation-lag.ts` is derived from git commit timestamps, not file content.
Regenerate it after every commit that touches those `.mdx` docs.

## 8. Milestones

Milestones live in `tasks/milestones/active/M-NNN-short-name/M-NNN.md`.

When executing a milestone:

- Read every file listed under Context before editing
- Stay inside Affected Modules and Out of Scope
- Follow Tasks in order; phases are sequential
- Use the branch named in the spec, usually `milestone/M-NNN-<slug>`
- Record blockers in the spec instead of widening scope silently
- Run the full acceptance criteria before pushing
- Package the PR with `docs/github-workflows/milestone-pr-template.md`

Scaffold helpers:

- `make new-milestone`
- `make new-plan`
- `make new-card`
- `make new-verify`
- `make run-card C=active/NNNN-slug`
- `make milestone-status M=active/M-NNN`

## 9. Anti-Patterns

Do not introduce:

- Stable package imports of `x/*`
- New dependencies without approval
- Context-based service lookup or package-level mutable registries
- `init()` side-effect registration
- Route auto-discovery or reflection-based wiring
- Middleware that builds business DTOs or injects services
- Ad hoc JSON response helpers or per-feature error envelopes
- New handler signatures
- New panic-only constructors for fallible behavior
- Generic `utils` packages
- Compatibility wrappers without a removal plan
