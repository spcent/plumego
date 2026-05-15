# CLAUDE.md - plumego

Go module: `github.com/spcent/plumego`
Go version: `go 1.24.0`, toolchain `go1.24.4`
Main module policy: standard library only unless explicitly approved.

Use `AGENTS.md` as the hard-rule entrypoint. For implementation detail, read the
control plane in this order:

1. `docs/CODEX_WORKFLOW.md` — operating modes and task recipes
2. `docs/CANONICAL_STYLE_GUIDE.md` — code shape and canonical examples
3. `docs/AGENT_CODE_QUALITY_RULES.md` — preflight checklist, gate selection, review output contract
4. `docs/agent-first.md` — agent-first design rationale and mechanism reference
5. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md` — repository layout
6. `docs/architecture/core-boundary.md` — stable root package responsibilities
7. `docs/architecture/extension-boundary.md` — x/* taxonomy, maturity ladder, promotion criteria
8. `specs/repo.yaml` — current stable and extension roots
9. `specs/task-routing.yaml` — task-to-entrypoint routing
10. `specs/extension-taxonomy.yaml` — primary and subordinate extension families
11. `specs/package-hotspots.yaml` — ambiguous package landing zones
12. `specs/dependency-rules.yaml` — import boundaries
13. `specs/agent-quality-rules.yaml` — machine-readable quality rules
14. target `<module>/module.yaml` — module-local scope and checks
15. `reference/standard-service` — canonical application wiring

---

## Repository Layout

### Stable Roots (long-lived, narrow public API)

| Package | Responsibility |
|---|---|
| `core` | App kernel: lifecycle, route registration, middleware attachment, server startup |
| `router` | Route matching, params, groups, reverse routing |
| `contract` | Transport primitives: response/error helpers, context accessors, binding |
| `middleware` | Transport-layer cross-cutting only (logging, recovery, timeout, CORS, auth adapters) |
| `security` | Auth primitives: JWT, password hashing, headers, input safety, abuse guard |
| `store` | Persistence primitives: DB, cache, file, KV, idempotency |
| `health` | Health models, readiness helpers |
| `log` | Logging contracts |
| `metrics` | Metrics contracts |

### Extension Families (`x/*`)

App-facing: `x/tenant`, `x/ai`, `x/rest`, `x/websocket`, `x/gateway`, `x/fileapi`,
`x/observability`, `x/messaging`, `x/frontend`, `x/resilience`, `x/ops`

Supporting: `x/cache`, `x/data`, `x/discovery`, `x/mq`, `x/pubsub`,
`x/scheduler`, `x/webhook`, `x/ipc`, `x/devtools`

Beta families (release-backed): `x/rest`, `x/websocket`, `x/gateway`, `x/observability`

When choosing a landing zone, start with `specs/task-routing.yaml` and use the
primary family entrypoint, not a subordinate package.

### Other Directories

| Path | Purpose |
|---|---|
| `reference/standard-service` | Canonical application layout (only canonical layout) |
| `reference/with-*` | Scenario-specific reference apps |
| `cmd/plumego/` | CLI: dev server, dashboard, scaffold |
| `internal/checks/` | Boundary, manifest, workflow, and maturity validation |
| `internal/testkit/` | Shared test helpers |
| `specs/` | Machine-readable authority: routing, taxonomy, dependency rules, quality rules |
| `specs/change-recipes/` | Reusable prompt recipes for standard task shapes |
| `tasks/milestones/` | Milestone specs and execution plans |
| `tasks/cards/` | Incremental task cards |
| `docs/` | All control-plane documentation |
| `website/` | Documentation generation |

---

## Operating Modes

### Analysis Mode

Use when scope, ownership, or architecture is unclear. Do not edit code.

Expected output: owning module, in-scope paths, out-of-scope paths, likely
touched files, main risks, validation plan.

### Implementation Mode

Use when the task contract is clear. Confirm owning module first, make the
smallest coherent change, add focused tests, run validation before handoff.

### Review Mode

Use for review, audit, or risk assessment. Do not patch unless explicitly asked.
Findings first, severity-ordered, with file references.

---

## Claude Defaults

### Non-Negotiables

- Preserve `net/http` compatibility.
- Keep the main module dependency-free beyond stdlib unless explicitly approved.
- Stable roots must not import `x/*`.
- No hidden globals, `init()` registration, or context service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth, verification, and policy errors.
- Use timing-safe comparison for secret checks.
- Context accessors use `With{Type}` + `{Type}FromContext`; key types are
  unexported zero-value structs inlined at the call site.
- `contract` owns transport primitives only.
- Deprecated symbols removed in the same PR that replaces their last caller.

### Code Shape

- Handler shape: `func(http.ResponseWriter, *http.Request)`
- Route wiring: one method, one path, one handler per line
- JSON decode: `json.NewDecoder(r.Body).Decode(...)`
- Error write: `contract.WriteError` with structured error codes
- Success write: `contract.WriteResponse`
- DI: constructor-based, visible in route wiring
- Middleware: `func(http.Handler) http.Handler`, transport-only, `next` called exactly once
- No service injection, business DTO assembly, or domain-policy branching in middleware
- Reference app: `reference/standard-service`

### Application Structure

```text
cmd/myservice/main.go
internal/httpapp/app.go
internal/httpapp/routes.go
internal/httpapp/handlers/<handler>.go
internal/domain/<domain>/service.go
internal/domain/<domain>/repository.go
```

Use constructor-based DI; wiring visible in the routes file. Do not mix routing,
domain logic, persistence, and transport policy in one package.

---

## Task Contract Defaults

When not stated explicitly, assume:

- one primary module per change
- no stable public API changes
- no new dependencies
- focused tests required for behavior changes
- docs sync only for implemented behavior, API, config, security, lifecycle, or
  boundary changes

### Stop Conditions

Stop in analysis mode (do not edit) when:

- owning module is unclear
- task would force a stable root to import `x/*`
- task needs a stable public API change that was not requested
- task needs a new dependency that was not approved
- task is broad but lacks acceptance criteria or validation commands
- a spec, module manifest, and local pattern conflict in a behavior-changing way

---

## Preflight Checklist

Complete before editing any file:

```text
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

---

## Symbol Change Protocol

When removing, renaming, or changing an exported symbol:

1. Enumerate all callers: `rg -n --glob '*.go' 'SymbolName' .`
2. Address every site in the same change (migrate, update, or explicit discard).
3. Re-run the same search — deletion leaves zero results; rename leaves no old name.
4. Update tests in the same commit.
5. `go build ./...` and required tests must pass before handoff.

---

## Validation

### Default Order

1. Run target module checks from `<module>/module.yaml`.
2. Run boundary and manifest checks.
3. Run repo-wide gates for code-bearing, cross-module, or release-relevant changes.

### Boundary and Manifest Checks

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
```

### Repo-Wide Gates

```bash
make gates
```

`make gates` mirrors CI: boundary checks, `go vet ./...`, non-mutating
`gofmt -l .`, race tests, and normal tests. Run `gofmt -w <paths>` before the
gate if formatting is needed.

### Gate Selection (lightest sufficient profile)

| Change type | Gate |
|---|---|
| Docs only | Check accuracy, links, authority order. No Go gates unless code/config changed. |
| Single module behavior | `go test -race ./<module>/...`, `go vet ./<module>/...` |
| Stable root change | Single-module gates + dependency-rules + agent-workflow + module-manifests |
| router / middleware / security | Module gates + dependency-rules; confirm negative-path and ordering tests |
| Cross-module, public API, release | Full `make gates` + all boundary checks + extension maturity/beta-evidence + `deprecation-inventory -strict` |

---

## Change Recipes

Before writing a bespoke prompt, check `specs/change-recipes/` for a matching
standard shape:

| Recipe | Task shape |
|---|---|
| `analysis-only.yaml` | Scope unknown; do not edit |
| `fix-bug.yaml` | General bug fix |
| `http-endpoint-bugfix.yaml` | Route/transport regression |
| `review-only.yaml` | Audit without patching |
| `add-http-endpoint.yaml` | New route + handler |
| `add-middleware.yaml` | New transport middleware |
| `new-stable-module.yaml` | New stable root package |
| `new-extension-module.yaml` | New x/* extension |
| `stable-root-boundary-review.yaml` | Stable root drift audit |
| `symbol-change.yaml` | Exported symbol rename/remove |
| `tenant-policy-change.yaml` | Tenant resolution, quota, isolation |

---

## Milestone Workflow

Milestones are human-scoped, single-PR execution units in
`tasks/milestones/active/M-NNN.md`.

When executing a milestone:

- read every file listed under Context before editing
- stay inside Affected Modules and Out of Scope
- follow Tasks in order; phase order is sequential
- use the branch named in the spec (`milestone/M-NNN-<slug>`)
- record blockers in the spec instead of widening scope silently
- run full acceptance criteria before pushing
- package the PR with `docs/github-workflows/milestone-pr-template.md`

Scaffold: `make new-milestone`, `make new-plan`, `make new-card`, `make new-verify`.
Use milestones for multi-step work; use daily implementation prompts for small fixes.

---

## Docs Sync Targets

Update these when behavior, public API, config, security semantics, lifecycle, or
boundaries change:

- `README.md`
- `README_CN.md`
- `AGENTS.md`
- `CLAUDE.md`
- `docs/ROADMAP.md`
- `env.example`

Document implemented behavior only.

---

## Anti-Patterns

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
