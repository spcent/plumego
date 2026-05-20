# AGENTS.md - plumego

Operational guide for AI coding agents working in `github.com/spcent/plumego`.

Plumego is an agent-first Go toolkit built on the standard-library HTTP model.
Keep the stable surface small, ownership obvious, and every change easy to
review and reverse.

Go module: `github.com/spcent/plumego`
Go version: `go 1.26.0` with Go toolchain `go1.26.0` or a newer patch release.
Main module policy: standard library only unless explicitly approved.

## 1. Authority

Use this file for hard rules and validation order. Read only the control-plane
files needed for the task:

- Workflow: `docs/CODEX_WORKFLOW.md`, `docs/AGENT_CODE_QUALITY_RULES.md`,
  `docs/AGENT_CONTEXT_BUDGET.md`, `specs/agent-quality-rules.yaml`, and
  `specs/change-recipes/`
- Style: `docs/CANONICAL_STYLE_GUIDE.md`
- Architecture: `docs/agent-first.md`,
  `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`,
  `docs/architecture/core-boundary.md`,
  `docs/architecture/extension-boundary.md`, and
  `docs/EXTENSION_MATURITY.md`
- Machine rules: `specs/repo.yaml`, `specs/task-routing.yaml`,
  `specs/extension-taxonomy.yaml`, `specs/extension-maturity.yaml`,
  `specs/package-hotspots.yaml`, `specs/dependency-rules.yaml`,
  `specs/checks.yaml`, `specs/module-manifest.schema.yaml`,
  `specs/stop-condition-handlers.yaml`, and `specs/request-flows.yaml`
- Local scope: target `<module>/module.yaml`
- Canonical app wiring: `reference/standard-service`

Default to the smallest safe context package. Read `AGENTS.md`, select the
matching `specs/task-routing.yaml` entry, then load only that entry's
`start_with` files and the target `<module>/module.yaml` when module behavior
changes. Use the full control plane only for architecture, boundary, release, or
workflow-rule changes.

When guidance conflicts, follow this order:

1. security and boundary rules in this file
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. machine-readable specs
4. existing local patterns in touched files

## 2. Non-Negotiables

- Preserve `net/http` compatibility.
- Keep the main module dependency-free beyond the standard library unless
  explicitly approved.
- Do not add `go.mod` anywhere under `x/**`; extension packages are part of the
  main module. Optional third-party adapters belong in reference apps or
  external modules.
- Stable roots must not import `x/*`.
- Do not introduce hidden globals, `init()` registration, or context
  service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth, verification, and policy errors.
- Use timing-safe comparison for secret checks.
- Context accessors use `With{Type}` plus `{Type}FromContext`; key types are
  unexported zero-value structs inlined at the call site.
- `contract` owns transport primitives only: response/error helpers, request
  metadata, context accessors, and binding helpers.
- Deprecated symbols must be removed in the same PR that replaces their last
  caller. Do not leave dead wrappers behind.
- Keep one canonical success-response path and one canonical error-construction
  path per layer.

## 2.1 Context Budget

- Use `docs/AGENT_CONTEXT_BUDGET.md` for context package selection, task-card
  limits, output compression, and resume discipline.
- Prefer the matching `specs/task-routing.yaml` `start_with` list over the full
  control-plane read order.
- Stop reading when ownership, boundaries, touched files, and validation are
  clear.
- Split broad work before implementation when it spans more than one primary
  module, more than five files, more than three validation commands, or unclear
  API/dependency/security impact.
- Summarize validation with command, status, key failure, and next step instead
  of pasting full logs.

## 3. Where To Work

Stable roots are long-lived public packages:

- `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`,
  `log`, `metrics`

Current top-level extension roots are:

- `x/ai`, `x/data`, `x/fileapi`, `x/frontend`, `x/gateway`, `x/messaging`,
  `x/observability`, `x/resilience`, `x/rest`, `x/tenant`, `x/websocket`

Subordinate extension packages live under those roots, for example
`x/data/cache`, `x/gateway/discovery`, `x/messaging/pubsub`, and
`x/observability/devtools`. Use `specs/task-routing.yaml` and
`specs/extension-taxonomy.yaml` to start at the primary family unless the task
is already narrowed to a subordinate package.

Default routing:

- kernel, lifecycle, routing, transport contracts, transport middleware, auth
  primitives, storage primitives: stable root
- product capability, protocol adaptation, business feature, extension behavior:
  `x/*`
- application wiring, bootstrap, DI, route registration:
  `reference/standard-service` or `cmd/plumego/internal/scaffold`
- architecture rules, boundaries, quality gates: `specs/`
- execution plans and task sequencing: `tasks/cards/` or `tasks/milestones/`

Hard boundaries:

- `core` is the app kernel, not a feature catalog.
- `router` owns matching, params, groups, and reverse routing.
- `middleware` stays transport-only; no business DTO assembly or service
  injection.
- `health` owns health models/readiness helpers, not HTTP handler ownership.
- Tenant logic belongs in `x/tenant`, not stable `middleware` or `store`.
- Tracing, sampling, collectors, and metric export wiring belong in
  `x/observability`, not `contract`.
- Session lifecycle belongs in `x/tenant` or an owning extension, not
  `contract`.
- New library code must live under a stable root or `x/*`; avoid new broad
  top-level roots such as `net`, `utils`, `validator`, `tenant`, `ai`, `rest`,
  `frontend`, or `pubsub`.

### Other Directories

| Path | Purpose |
|---|---|
| `reference/standard-service` | Canonical application layout |
| `reference/production-service` | Production-oriented explicit wiring reference |
| `reference/with-*` | Scenario-specific reference apps |
| `reference/workerfleet` | Separate worker-oriented submodule reference |
| `cmd/plumego/` | CLI: dev server, dashboard, scaffold |
| `internal/checks/` | Boundary, manifest, workflow, and maturity validation |
| `internal/testutil/` | Shared test helpers |
| `specs/` | Machine-readable authority: routing, taxonomy, dependency rules, quality rules |
| `specs/change-recipes/` | Reusable prompt recipes for standard task shapes |
| `specs/stop-condition-handlers.yaml` | Resolution paths for every agent stop condition |
| `specs/request-flows.yaml` | Machine-readable request flow diagrams (module ownership per step) |
| `tasks/milestones/` | Milestone specs and execution plans |
| `tasks/cards/` | Incremental task cards |
| `docs/` | All control-plane documentation |
| `website/` | Documentation generation |

## 4. Canonical Defaults

- Handler shape: `func(http.ResponseWriter, *http.Request)`
- Route wiring: one method, one path, one handler per line
- JSON decode: `json.NewDecoder(r.Body).Decode(...)`
- Error write path: `contract.WriteError` with structured error codes
- Success write path: `contract.WriteResponse`
- DI: constructor-based and visible in route wiring
- Middleware: `func(http.Handler) http.Handler`, transport-only; `next` called exactly once; no service injection, business DTO assembly, or domain-policy branching
- Reference app: `reference/standard-service`

Do not add one-off handler styles, response envelopes, error helper families, or
route registration idioms.

## 5. Working Contract

For normal implementation requests, assume:

- one primary module per change
- no stable public API changes
- no new dependencies
- focused tests for behavior changes
- docs sync only for behavior, API, config, security, lifecycle, or boundary
  changes

Use analysis mode and do not edit when ownership is unclear, the task is broad
without acceptance criteria, a new dependency is required, or a stable public API
change was not requested. When you hit a stop condition, read
`specs/stop-condition-handlers.yaml` for the deterministic resolution path —
each condition maps to concrete next steps, an escalation action, and the
unblock criteria.

For review requests, do not patch unless explicitly asked. Findings come first,
ordered by severity, with boundary violations, regressions, hidden coupling, and
missing tests prioritized.

Before editing, complete the following preflight checklist (full rules in
`docs/AGENT_CODE_QUALITY_RULES.md`):

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

## 6. Change Rules

- Keep changes minimal and scoped to one primary module when possible.
- Read the target module manifest before editing module behavior.
- Preserve stable public APIs unless explicitly asked to change them.
- If a breaking change is unavoidable, include migration notes.
- Prefer standard-library solutions and existing local patterns.
- Add or update tests next to changed behavior.
- Do not refactor unrelated files opportunistically.

## 7. Symbol Changes

### 7.1 Completeness Protocol for Symbol Changes

When removing, renaming, or changing the behavior of an exported symbol:

1. Enumerate all sites first with `rg -n --glob '*.go' 'SymbolName' .`.
2. Address every site in the same change: migrate, update assertions, or add an
   explicit `_ =` discard with a TODO.
3. Re-run the same search. For deletion, the result must be empty. For rename,
   the old name must not appear.
4. Update tests in the same commit.
5. Before handoff, `go build ./...` and the required tests must pass.

No caller may silently discard a newly returning function without an explicit
discard.

## 8. Validation

Default order:

1. Run target module checks from `<module>/module.yaml`.
2. Run boundary and manifest checks.
3. Run repo-wide gates when the change is code-bearing, cross-module, or release
   relevant.

Boundary and manifest checks:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/cross-extension-deps
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go run ./internal/checks/public-entrypoints-sync
```

Repo-wide gates:

```bash
make gates
```

`make gates` mirrors CI: boundary checks, `go vet ./...`, a non-mutating
`gofmt -l .` check, race tests, and normal tests. Run `gofmt -w <paths>` before
the gate if formatting is needed.

Extra focus:

- routing: static, params, groups, and reverse routing
- middleware: ordering and error paths
- security: invalid token/signature negative tests
- tenant: quota, policy, and isolation
- store: concurrency and persistence correctness

Select the lightest sufficient gate profile:

| Change type | Gate |
|---|---|
| Docs only | Check accuracy, links, authority order. No Go gates unless code/config changed. |
| Single module behavior | `go test -race ./<module>/...`, `go vet ./<module>/...` |
| Stable root change | Single-module gates + dependency-rules + agent-workflow + module-manifests |
| router / middleware / security | Module gates + dependency-rules; confirm negative-path and ordering tests |
| Cross-module, public API, release | Full `make gates` + extension maturity/beta-evidence + `deprecation-inventory -strict` |

See `specs/agent-quality-rules.yaml` for the machine-readable version of these profiles.

Do not expand checked-in baselines or generated snapshots casually.

## 9. Docs Sync

Update affected docs only when behavior, public API, config, security
semantics, lifecycle behavior, or boundaries change. Common sync targets:

- `README.md`
- `README_CN.md`
- `AGENTS.md`
- `docs/AGENT_CONTEXT_BUDGET.md`
- `env.example`

Also update module primers under `docs/modules/`, `docs/EXTENSION_MATURITY.md`,
`docs/ROADMAP.md`, or `docs/stable-api/` when the change specifically affects
that surface.

Document implemented behavior only.

### Website Generated Files

`website/src/generated/` contains files generated from docs sources. These
files are checked in and must stay in sync. `make gates` detects staleness by
running the sync and comparing before/after diffs; it errors if the sync
produces any change.

**Required when editing any of these sources:**

| Source file | Generated file | Rule |
|---|---|---|
| `docs/ROADMAP.md` | `website/src/generated/roadmap.ts` | run `make website-sync` |
| `docs/modules/*/README.md` | `website/src/generated/modules.ts` | run `make website-sync` |
| `specs/task-routing.yaml` | `website/src/generated/routing.ts` | run `make website-sync` |
| `docs/release/*.md` | `website/src/generated/release-meta.ts` | run `make website-sync` |

After any edit to a source listed above, run `make website-sync` and include
the updated `website/src/generated/` files in the same commit. Never commit
source changes without the corresponding generated file update.

## 10. Milestones

Milestones are human-scoped, single-PR execution units in
`tasks/milestones/active/M-NNN-short-name/M-NNN.md`.

When executing a milestone:

- read every file listed under Context before editing
- stay inside Affected Modules and Out of Scope
- follow Tasks in order; phase order is always sequential
- use the branch named in the spec, usually `milestone/M-NNN-<slug>`
- record blockers in the spec instead of widening scope silently
- run the full acceptance criteria before pushing
- package the PR with `docs/github-workflows/milestone-pr-template.md`

Scaffold commands: `make new-milestone`, `make new-plan`, `make new-card`,
`make new-verify`. Use milestones for multi-step work; use task cards for small
fixes.

Additional targets:

- `make run-card C=active/NNNN-slug` — validate card, generate bundle, launch execution
- `make milestone-status M=active/M-NNN` — display phase checkpoint progress

See `docs/MILESTONE_PIPELINE.md` and `tasks/milestones/README.md` for detailed
planning, card, verify, and PR packaging rules.

## 11. Anti-Patterns

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
- New broad root names: `ai`, `tenant`, `net`, `pubsub`, `rest`, `validator`,
  `utils`, `frontend`
