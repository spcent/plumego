# Core Package Refactor Plan (Canonical-Only, No Compatibility Constraints)

Status: Proposed  
Scope: `core/` package and directly coupled entrypoints  
Primary reference: `docs/CANONICAL_STYLE_GUIDE.md`

---

## 1. Objective

Refactor `core/` into a **single canonical runtime surface** that is:

- standard-library-first (`http.Handler`, `http.HandlerFunc`),
- explicit in lifecycle and wiring,
- narrow in responsibility (runtime/lifecycle only),
- easier to reason about for humans and AI agents.

This plan intentionally does **not** preserve compatibility APIs. The target is a clean v1-style core surface aligned with the canonical guide.

---

## 2. Canonical Principles Applied to `core/`

1. **One route registration style** in `core.App`: method + path + `http.HandlerFunc`.
2. **No compatibility handler families** in `core` (`*Ctx`, `*Handler`, or similar alternate surfaces).
3. **No hidden dependency lookup** in request path; dependencies are constructor-injected at app wiring boundaries.
4. **Middleware is transport-only** and remains outside business/domain concerns.
5. **Core is runtime orchestration only**; feature catalogs and business helpers stay in extension packages.
6. **Small explicit lifecycle pipeline**: construct → configure middleware/routes/components → build server → run/shutdown.

---

## 3. Current Core Pain Points (Refactor Drivers)

### 3.1 API Surface Overexpansion

`core/routing.go` exposes canonical methods (`Get/Post/...`) and compatibility variants (`GetCtx/PostCtx/...`, `GetHandler/PostHandler/...`). This widens the public API and increases maintenance/test burden.

### 3.2 Monolithic Lifecycle Flow

`core/lifecycle.go` contains a long end-to-end startup/shutdown path with many responsibilities in one file (env loading, component wiring, DI registration/injection, server setup, signal lifecycle, shutdown sequencing).

### 3.3 Option Surface Too Broad for Core Boundary

`core/options.go` mixes pure runtime options with capability toggles and feature wiring concerns. This makes `core.New(...)` look like a feature gateway instead of a small runtime bootstrap API.

### 3.4 Hidden Wiring Paths

Core-level DI registration and injection inside startup (`app.go` + `lifecycle.go`) reduce wiring visibility. For canonical style, dependency ownership should be obvious at construction boundaries.

### 3.5 Wrapper Proliferation

Multiple wrapper files (`*_wrapper.go`) increase indirection for feature mounting. Core orchestration is harder to follow because feature behavior is fragmented across wrappers.

### 3.6 Test Concentration and Fragility Risk

Large test files (`lifecycle_test.go`, `options_test.go`, `app_test.go`) suggest broad coupling and make safe iteration harder.

---

## 4. Target Architecture

## 4.1 Public API Target (`core`)

Keep only a compact canonical API:

- Construction: `core.New(opts ...Option)`
- Middleware: `app.Use(mw ...func(http.Handler) http.Handler)`
- Routes: `app.Get/Post/Put/Delete/Patch/Any(path, handler)`
- Serving:
  - `app.ServeHTTP(w, r)` (already handler-compatible)
  - `app.Server()` or `app.BuildServer()` for explicit `http.Server` ownership
  - `app.Run(ctx)` for managed lifecycle (optional but single canonical runtime entry)

Remove from core public API:

- `GetCtx/PostCtx/...`
- `GetHandler/PostHandler/...`
- context-adapter registration helpers as public first-class methods

## 4.2 Internal Package Layout Target

Refactor `core/` internals into explicit runtime subdomains:

- `core/runtime/boot.go` — startup orchestration
- `core/runtime/server.go` — `http.Server` assembly and listen logic
- `core/runtime/shutdown.go` — signal handling + graceful stop
- `core/runtime/components.go` — component mount/start/stop sequencing
- `core/runtime/state.go` — immutable config + mutable runtime state boundaries
- `core/runtime/guards.go` — mutability/freeze checks

Keep `core/app.go` minimal as a façade over runtime modules.

## 4.3 Option Model Target

Split options into two tiers:

1. **Core runtime options** (address, timeouts, TLS, shutdown, logger)
2. **Extension/component options** managed by component constructors or dedicated feature packages

Rule: if an option configures a non-runtime capability (for example webhook/pubsub devtools behavior), it should not live in core’s primary option set.

## 4.4 Component Model Target

- Keep `Component` focused on route/middleware registration and lifecycle.
- Remove reflection-heavy dependency declarations from component contracts in favor of explicit constructor wiring order.
- Resolve startup ordering through explicit dependency graph passed by app builder (not implicit runtime injection).

## 4.5 Error and Observability Target

- Core runtime errors should be structured and stable.
- One runtime error path for start/stop failures (clear categories: config, wiring, listen, shutdown).
- Logging should describe lifecycle transitions but never hide control flow.

---

## 5. Work Breakdown Structure (WBS)

## Phase 0 — Baseline and Guardrails (1 PR)

1. Freeze desired canonical API contract in a design doc.
2. Add architecture decision record (ADR): “core is canonical runtime only.”
3. Add API-surface test that fails when new non-canonical route registration methods appear.

Deliverable: approved API boundary before code movement.

## Phase 1 — Route API Contraction (2–3 PRs)

1. Remove `*Ctx` and `*Handler` route methods from `core.App`.
2. Keep only `http.HandlerFunc` registration methods in `core/routing.go`.
3. Update all in-repo callers under `examples/` and docs to canonical handlers.
4. Simplify routing tests to canonical-only method tables.

Deliverable: single route registration surface in core.

## Phase 2 — Lifecycle Decomposition (3–5 PRs)

1. Split `core/lifecycle.go` into runtime modules (`boot/server/shutdown/components`).
2. Extract signal handling into isolated unit-tested shutdown coordinator.
3. Make startup sequence explicit and linear:
   - validate config
   - build handler chain
   - build server
   - start components/runners
   - listen
4. Ensure every stage has deterministic rollback on failure.

Deliverable: readable lifecycle graph with isolated tests per stage.

## Phase 3 — Option Surface Simplification (2–4 PRs)

1. Classify all `WithXxx` options into runtime vs extension.
2. Move non-runtime feature options out of core into component builders.
3. Keep `core.New(...)` examples minimal and canonical.
4. Replace option side-effects with explicit builder wiring where possible.

Deliverable: smaller, runtime-focused `core/options.go`.

## Phase 4 — Component Wiring Explicitness (2–4 PRs)

1. Replace implicit DI-based component injection with explicit constructor dependencies.
2. Remove runtime container registrations that are not essential to core lifecycle.
3. Enforce that route and middleware ownership is visible from app assembly code.

Deliverable: no hidden dependency resolution in request/runtime path.

## Phase 5 — Wrapper Consolidation (2 PRs)

1. Consolidate `*_wrapper.go` files behind clear component adapters or remove where unnecessary.
2. Keep one obvious registration path per capability.
3. Ensure wrappers do not mutate global state or hide registration order.

Deliverable: lower indirection and grep-friendly feature wiring.

## Phase 6 — Test Refactor and Hardening (ongoing)

1. Split large test files by behavior area (boot, server config, shutdown, routing, options).
2. Add deterministic tests for startup rollback and shutdown ordering.
3. Add race tests for mutable runtime state transitions.
4. Add boundary tests ensuring core does not absorb extension behavior.

Deliverable: modular test suite aligned with runtime modules.

---

## 6. File-Level Change Plan

## 6.1 `core/routing.go`

- Keep only canonical `http.HandlerFunc` registration API.
- Remove compatibility adapters and related helper functions.
- Keep nil-handler and mutability checks centralized.

## 6.2 `core/lifecycle.go`

- Decompose into runtime-focused files under `core/runtime/`.
- Minimize cross-cutting responsibilities in one function.
- Keep `Boot`/`Run` façade short and explicit.

## 6.3 `core/options.go`

- Keep only runtime options in core.
- Move capability-specific options to feature/component packages.
- Eliminate options that trigger hidden side effects.

## 6.4 `core/app.go` and `core/app_helpers.go`

- Keep `App` as a thin state container + façade.
- Move state transitions into small internal services.
- Keep lock boundaries short and visible.

## 6.5 `core/components_default.go` and wrapper files

- Replace implicit built-in mounting with explicit opt-in component composition.
- If defaults remain, make them a single explicit preset option with documented composition.

## 6.6 `core/di/*`

- Restrict or remove DI usage from core runtime orchestration.
- Keep DI (if retained) outside request path and outside hidden lifecycle auto-wiring.

---

## 7. Testing Strategy and Quality Gates

Required gates for every phase:

```bash
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

Additional core-focused checks:

1. Routing tests: method/path dispatch, group behavior, params, reverse routing interaction.
2. Middleware tests: ordering invariants and error-path behavior after lifecycle refactor.
3. Lifecycle tests:
   - startup failure rollback (components/runners/server)
   - signal-triggered graceful shutdown
   - shutdown timeout and drain behavior
4. Concurrency tests:
   - freeze/mutability guard races
   - handler initialization once semantics

---

## 8. Execution Order and Milestones

## Milestone A — Canonical API lock

- Route API contraction complete.
- Docs/examples no longer use removed core compatibility methods.

## Milestone B — Lifecycle clarity

- Runtime decomposition merged.
- Startup/shutdown behavior unchanged semantically but easier to trace.

## Milestone C — Boundary enforcement

- Option surface reduced to runtime concerns.
- Extension wiring moved out of core.

## Milestone D — Maintainability baseline

- Test suites modularized and stable.
- Core package complexity reduced (file sizes, coupling, and public API count).

---

## 9. Risk Register

1. **Large-scope breakage across examples/tests**
   - Mitigation: phase-by-phase PRs with repo-wide caller updates in each phase.
2. **Lifecycle regressions during decomposition**
   - Mitigation: snapshot tests before refactor + failure rollback tests after each extraction.
3. **Feature behavior drift while moving options/wrappers**
   - Mitigation: contract tests at component boundaries and explicit migration notes.
4. **Concurrency regressions**
   - Mitigation: race-focused tests around state transitions and startup/shutdown paths.

---

## 10. Definition of Done (DoD)

The refactor is complete when all conditions are met:

1. Core exposes exactly one canonical routing/handler style.
2. Core lifecycle path is split into small runtime modules with explicit sequencing.
3. Core option set is runtime-only; capability options are outside core.
4. No hidden DI/service locator pattern is required for normal core runtime behavior.
5. Docs and examples show canonical core bootstrap and route style only.
6. Quality gates and core-specific tests pass consistently.

---

## 11. Suggested PR Slicing Template

Use this recurring template for each PR in the program:

1. **Intent**: one sentence
2. **Scope**: files/modules touched
3. **Behavioral changes**: explicit list
4. **Risk**: what could break
5. **Verification**: exact test commands + key test cases

Keep each PR reversible and focused on one phase objective.
