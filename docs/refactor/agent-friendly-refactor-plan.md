# Plumego Agent-Friendly Refactor Plan

Status: proposed  
Branch: `claude/plumego-agent-refactor-plan-O11PN`  
Scope: analysis + planning only — no code changes in this document

---

## 1. Summary

Plumego already has a sophisticated agent-first infrastructure: `AGENTS.md`, layered
`specs/` YAML, per-module `module.yaml` manifests, `internal/checks/` scripts, milestone
and card scaffolding, and `specs/change-recipes/`. This is better than most open-source
Go repositories. The gaps are not about structure from scratch — they are about filling
five missing layers and fixing a handful of concrete violations that cause Agent confusion
and rework.

**The five gaps:**
1. No `agents/` working-rules directory — agents must synthesize rules from 12+ scattered files
2. No `specs/public-api.yaml` — no machine-readable contract for what's stable vs. internal
3. `make gates` requires a website build — creates friction for Go-only changes
4. Ten undeclared `internal/` usages in `x/` not listed in `dependency-rules.yaml`
5. `contract.RouteStatePool` is an exported mutable global — violates the written non-negotiables

---

## 2. Current Problems

### 2.1 What Works Well

- `specs/dependency-rules.yaml` with an automated check
- Per-module `module.yaml` for all stable roots and most extensions
- `specs/change-recipes/` gives agents deterministic workflows per task shape
- `specs/stop-condition-handlers.yaml` gives structured off-ramps for blocks
- `internal/checks/` scripts are executable gates, not prose

### 2.2 Agent-Hostile Patterns

**P0: No single working-rules file for agents**

`AGENTS.md` is 330+ lines and cross-references 12+ files that agents must read before
acting. There is no machine-readable "what am I allowed to touch in this module in this
task shape" file. Agents synthesize from AGENTS.md → specs/repo.yaml → specs/task-routing.yaml
→ module.yaml → specs/dependency-rules.yaml. Any one misread compounds into rework.

**P1: `make gates` is indivisible and heavyweight**

`make gates` runs Go tests, boundary checks, reference tests, stable-API snapshots,
doc-snippets, `gofmt`, `go vet`, **and** `pnpm sync + pnpm check + pnpm build` of the website.
A pure Go change cannot pass gates without a working Node environment. Agents cannot
validate locally without the website toolchain.

**P1: `contract.RouteStatePool` violates the written non-negotiable**

`contract/context_core.go:56` exports `RouteStatePool = sync.Pool{...}`. AGENTS.md §2
says "Do not introduce hidden globals." This is exported mutable shared state in a stable
package — a concrete violation any agent reading the rules would flag, and yet it's merged.

**P1: `internal/` package usages in `x/` are undeclared in `dependency-rules.yaml`**

Ten imports of `internal/` packages from `x/` packages are not declared anywhere:

| x/ package | internal/ package used |
|---|---|
| `x/ai/marketplace` | `internal/semver` |
| `x/gateway/cache`, `x/gateway/transform` | `internal/httputil` |
| `x/gateway/websocket` | `internal/httpx` |
| `x/messaging/provider_http.go` | `internal/nethttp` |
| `x/messaging/webhook` | `internal/httpx`, `internal/stringsx` |
| `x/observability/devtools` | `internal/config` |

An agent checking the rules would think these are violations; a checking agent would
not know they are intentional.

**P1: `internal/config` panics on nil logger**

`internal/config/manager.go:NewManager()` panics when logger is nil.
AGENTS.md §2 says "No new panic-only constructors for fallible behavior."

**P2: No `specs/public-api.yaml`**

There is a `docs/stable-api/snapshots/` directory and a checker tool, but no
machine-readable file listing the stable public API surface with status tags. Agents
cannot quickly answer "is this symbol GA or experimental?" without reading multiple manifests.

**P2: No `agents/` directory**

The five agent-working-rules files (`working-rules.md`, `safe-refactor-zones.md`,
`verification-matrix.md`, `prompt-recipes.md`, `rollback-playbook.md`) do not exist.

**P2: No `module.yaml` for `internal/` packages**

`internal/config`, `internal/httputil`, `internal/httpx`, `internal/nethttp`,
`internal/semver`, `internal/stringsx`, `internal/testutil` all lack `module.yaml`.

**P2: Test naming is inconsistent**

`freeze_test.go`, `conformance_test.go`, `_coverage_test.go`, `_extended_test.go` do not
declare their purpose in a scannable way.

**P3: No `adr/` directory, no `examples/` directory**

Architectural decisions are not recorded. There are no minimal snippet-level examples
(only full reference apps).

---

## 3. Target Repository Structure (additive only)

```
plumego/
│
├── agents/                            # NEW: deduplicated agent working rules
│   ├── working-rules.md               # 1-page distillation of AGENTS.md non-negotiables
│   ├── safe-refactor-zones.md         # Per-package: what's safe vs. needs review
│   ├── verification-matrix.md         # Change type → exact gate commands
│   ├── prompt-recipes.md              # Copy-paste task prompts for common shapes
│   └── rollback-playbook.md           # How to revert each change type cleanly
│
├── adr/                               # NEW: architectural decision records
│   ├── README.md
│   ├── 001-stdlib-only-main-module.md
│   ├── 002-contract-owns-route-state.md
│   ├── 003-x-prefix-extension-convention.md
│   └── 004-internal-package-visibility.md
│
├── examples/                          # NEW: minimal runnable snippets
│   ├── hello-world/
│   ├── add-middleware/
│   ├── structured-error/
│   ├── route-params/
│   └── health-check/
│
├── specs/                             # Keep all; add:
│   ├── public-api.yaml                # NEW: machine-readable stable public API surface
│   ├── dependency-rules.yaml          # Fix: declare all internal/* usages
│   └── checks.yaml                    # Fix: add go_only and website_only profiles
│
├── internal/                          # Add module.yaml to each package
│   ├── config/module.yaml             # NEW
│   ├── httputil/module.yaml           # NEW
│   ├── httpx/module.yaml              # NEW
│   ├── nethttp/module.yaml            # NEW
│   ├── semver/module.yaml             # NEW
│   ├── stringsx/module.yaml           # NEW
│   └── testutil/module.yaml           # NEW
│
├── contract/context_core.go           # Fix: RouteStatePool → routeStatePool
└── Makefile                           # Add composable Go-only targets
```

---

## 4. Core Package Architecture — Delta from Current

The stable package architecture is well-designed. Only changes from current state:

### `contract`

**Fix:** `RouteStatePool` (exported `sync.Pool`) must become `routeStatePool` (unexported).
Router accesses it via an unexported accessor function added to `contract`.
`freeze_test.go` must be updated to remove `RouteStatePool` from the expected export list.

### `internal/config`

**Fix:** `NewManager(logger)` panics on nil logger. Change to `NewManager(logger) (*Manager, error)`.
Update the one caller in `x/observability/devtools/devtools.go`.

### All `internal/` packages

**Fix:** Add `module.yaml` declaring `layer: internal`, `status: internal`, and `declared_callers`.

### `specs/dependency-rules.yaml`

**Fix:** Add entries for `internal/semver`, `internal/httputil`, `internal/httpx`,
`internal/nethttp`, `internal/stringsx`, `internal/config` with their `declared_callers` lists.

---

## 5. Agent Control Plane — Five New Files

### `agents/working-rules.md`

One-page distillation of every AGENTS.md rule that applies to a typical change.
Agents read this first; AGENTS.md for exceptions.

Core content:
- 7 non-negotiables in bullet form (from AGENTS.md §2)
- 6-field preflight checklist
- "Before you edit, read: [module.yaml, dependency-rules.yaml section]"
- 4-bullet stop-criteria
- One-line gate-selection decision tree

### `agents/safe-refactor-zones.md`

Per-package table: Safe (low-risk) / Review Required (medium) / Human Required (high).

Example row:
```
| contract | error message text, new test | new exported type, new error code | remove any exported symbol |
```

### `agents/verification-matrix.md`

Exact `go` commands for each change type, replacing gate-selection prose with a lookup table.

```
| Change type               | Commands |
|---------------------------|----------|
| docs only                 | (none) |
| single stable-root file   | go test -race ./[pkg]/... && go vet ./[pkg]/... && gofmt -l . |
| new middleware sub-package| + go run ./internal/checks/dependency-rules |
| x/* behavior change       | go test -race ./x/[family]/... + dependency-rules |
| cross-module stable       | module gates + all boundary checks |
| public API change         | make gates-go (full) |
| new dependency            | STOP — requires human approval |
```

### `agents/prompt-recipes.md`

≥6 copy-paste prompt templates for common task shapes, each with pre-filled preflight.
Task shapes: add HTTP endpoint, fix middleware bug, add extension feature,
add security primitive, update module manifest, write contract test.

### `agents/rollback-playbook.md`

Deterministic rollback steps per change type with verification command.

```
| new exported symbol | rg -n 'Symbol' . → remove all → run freeze_test.go |
| new middleware      | remove from chain → delete file → conformance test |
| new module.yaml field | remove field → module-manifests check |
| new dependency      | go mod tidy → git diff go.sum must be empty |
```

---

## 6. Machine-Readable Specs Changes

### `specs/dependency-rules.yaml`

Add 6 new entries:

```yaml
  internal/semver:
    path: internal/semver
    layer: internal
    declared_callers:
      - x/ai/marketplace
    deny: [reference/**, cmd/**]

  internal/httputil:
    path: internal/httputil
    layer: internal
    declared_callers:
      - middleware/internal/transport
      - x/gateway/cache
      - x/gateway/transform
    deny: [reference/**, cmd/**]

  internal/httpx:
    path: internal/httpx
    layer: internal
    declared_callers:
      - x/gateway
      - x/messaging/webhook
    deny: [reference/**, cmd/**]

  internal/nethttp:
    path: internal/nethttp
    layer: internal
    declared_callers:
      - x/messaging
    deny: [reference/**, cmd/**]

  internal/stringsx:
    path: internal/stringsx
    layer: internal
    declared_callers:
      - x/messaging/webhook
    deny: [reference/**, cmd/**]

  internal/config:
    path: internal/config
    layer: internal
    declared_callers:
      - x/observability/devtools
    deny: [reference/**, cmd/**]
```

### `specs/checks.yaml`

Add profiles:

```yaml
  go_only:
    boundary_checks:
      - go run ./internal/checks/dependency-rules
      - go run ./internal/checks/agent-workflow
      - go run ./internal/checks/module-manifests
      - go run ./internal/checks/reference-layout
      - go run ./internal/checks/public-entrypoints-sync
    repo_wide_checks:
      - go vet ./...
      - gofmt -l .
      - go test -race -timeout 60s ./...
  website_only:
    - cd website && pnpm sync
    - cd website && pnpm check
    - cd website && pnpm build
```

### `specs/public-api.yaml` (new)

```yaml
version: 1
# Authoritative list of stable GA exported symbols.
# Checked by internal/checks/public-entrypoints-sync.
# Adding: declare here AND in module.yaml:public_entrypoints.
# Removing: remove all callers first, then remove from here AND module.yaml.

stable_surface:
  contract:
    status: ga
    breaking_change_policy: semver
    exported_symbols:
      - WriteResponse
      - WriteError
      - RequestContextFromContext
      - RequestParamFromContext
      - WithRequestContext
      - WithRouteState
      - RouteState
      - RequestContext
      # ... populated from contract/module.yaml:public_entrypoints

  core:
    status: ga
    breaking_change_policy: semver
    exported_symbols:
      - New
      - App
      - AppDependencies
      - AppConfig
      - DefaultConfig
      # ... populated from core/module.yaml:public_entrypoints

  # ... one entry per stable root
```

---

## 7. Testing and Verification

### Test Naming Convention (target)

| Filename pattern | Intent |
|---|---|
| `foo_test.go` | Unit test for `foo.go` |
| `foo_bench_test.go` | Benchmark |
| `freeze_test.go` | API snapshot — exported set must not change |
| `contract_test.go` | JSON wire format or interface contract |
| `example_test.go` | Runnable `Example*` functions |
| `conformance_test.go` | Shape conformance (e.g., middleware signature) |

Rename: existing `_coverage_test.go` → inline or `_contract_test.go`. Existing `_extended_test.go` → inline.

### New Makefile Targets

```makefile
check-boundaries: ## Run boundary and manifest checks only
    go run ./internal/checks/dependency-rules
    go run ./internal/checks/agent-workflow
    go run ./internal/checks/module-manifests
    go run ./internal/checks/reference-layout
    go run ./internal/checks/public-entrypoints-sync

check-public-api: ## Check stable public API surface hasn't drifted
    go run ./internal/checks/public-entrypoints-sync

test-stable: ## Run tests for stable roots only
    go test -race -timeout 60s ./core/... ./router/... ./contract/... \
      ./middleware/... ./security/... ./store/... ./health/... ./log/... ./metrics/...

test-contract: ## Run freeze/conformance/contract tests only
    go test -run 'TestFreeze|TestConformance|TestContract' -timeout 20s ./...

bench: ## Run all benchmarks (short mode)
    go test -bench=. -benchtime=1s -run=^$ ./...

gates-go: ## Run all Go quality gates (no website required)
    $(MAKE) check-boundaries
    $(MAKE) check-public-api
    go run ./internal/tools/doc-snippets
    go vet ./...
    @UNFORMATTED=$$(gofmt -l .); if [ -n "$$UNFORMATTED" ]; then echo "$$UNFORMATTED"; exit 1; fi
    go test -race -timeout 60s ./...
    $(MAKE) reference-test
    go run ./internal/checks/extension-maturity
    go run ./internal/checks/extension-beta-evidence
    go run ./internal/checks/deprecation-inventory -strict

gates-website: ## Run website quality gates
    cd website && pnpm sync
    cd website && pnpm check
    cd website && pnpm build
```

Modify existing `gates` target to call `$(MAKE) gates-go && $(MAKE) gates-website`.

---

## 8. Migration Roadmap

Each phase is a standalone, independently mergeable PR.

| Phase | Title | Scope | Definition of Done |
|---|---|---|---|
| 0 | Repository Audit | This document | Committed and pushed |
| 1 | Agent Control Plane | `agents/*.md` (5 new files) | Docs only; all 5 files complete |
| 2 | Specs Fixes | `dependency-rules.yaml`, `checks.yaml`, `public-api.yaml` | `dependency-rules` check passes |
| 3 | Internal Module Manifests | `internal/*/module.yaml` (7 new files) | `module-manifests` check passes |
| 4 | Makefile Targets | `Makefile` | `make gates-go` passes without pnpm |
| 5 | Fix RouteStatePool | `contract/`, `router/` | `RouteStatePool` unexported; freeze test passes |
| 6 | Fix config panic | `internal/config/`, `x/observability/devtools/` | `NewManager` returns error; tests pass |
| 7 | Examples | `examples/` (5 new dirs) | Each compiles; ≤50 lines; stdlib only |
| 8 | ADR | `adr/` (5 new files) | Docs only |
| 9 | Test Naming | rename `_coverage_test.go` files; add `freeze_test.go` to health/log/metrics | All tests pass |

---

## 9. First 10 PR Tasks

### PR 1: `agents/working-rules.md`

**Title:** `docs(agents): add working-rules distillation of AGENTS.md non-negotiables`  
**Goal:** Single file agents read first instead of traversing AGENTS.md + 4 companion docs.  
**Allowed files:** `agents/working-rules.md` (create only)  
**Forbidden files:** Any `.go` file; any existing `.md` file  
**Definition of done:** ≤120 lines; contains all 7 non-negotiables; 6-field preflight; gate-selection table  
**Required checks:** Docs only  
**Prompt for next agent:**
```
Create agents/working-rules.md. Must contain:
1. The 7 non-negotiables from AGENTS.md §2 in bullet form
2. A 6-field preflight checklist (owning-module, in-scope, out-of-scope, API-impact,
   security-impact, validation-plan)
3. Gate-selection table: docs-only → none; single-module → module gates; cross-module → full gates
4. Stop-condition list: unclear ownership, unapproved dependency, unapproved stable API change,
   boundary violation
File must be ≤120 lines. No code changes. No edits to existing files.
```

---

### PR 2: `agents/verification-matrix.md`

**Title:** `docs(agents): add verification-matrix — exact gate commands per change type`  
**Goal:** Replace prose gate-selection docs with a scannable table.  
**Allowed files:** `agents/verification-matrix.md` (create only)  
**Forbidden files:** Any `.go` file; `Makefile`; any existing spec file  
**Definition of done:** Table with ≥8 rows; exact shell commands; covers all change types from gate matrix in `specs/agent-quality-rules.yaml`  
**Required checks:** Docs only

---

### PR 3: `agents/safe-refactor-zones.md`

**Title:** `docs(agents): add safe-refactor-zones — per-package risk table`  
**Goal:** Per-package lookup: safe / needs-review / needs-human-approval.  
**Allowed files:** `agents/safe-refactor-zones.md` (create only)  
**Forbidden files:** Any `.go` file; any existing `.md` file  
**Definition of done:** Row per stable root; 3 risk columns; specific examples in each cell based on `module.yaml:change_risks`  
**Required checks:** Docs only

---

### PR 4: `agents/prompt-recipes.md`

**Title:** `docs(agents): add prompt-recipes — copy-paste task prompts`  
**Allowed files:** `agents/prompt-recipes.md` (create only)  
**Definition of done:** ≥6 copy-paste templates; each includes full pre-filled preflight checklist  
**Required checks:** Docs only

---

### PR 5: `agents/rollback-playbook.md`

**Title:** `docs(agents): add rollback-playbook — deterministic revert steps per change type`  
**Allowed files:** `agents/rollback-playbook.md` (create only)  
**Definition of done:** Table with rollback steps + verification command per change type; covers ≥6 change types  
**Required checks:** Docs only

---

### PR 6: Declare internal package usages in `dependency-rules.yaml`

**Title:** `specs: declare internal/* usages in dependency-rules.yaml`  
**Goal:** Convert 10 silent assumptions into verified contracts.  
**Allowed files:** `specs/dependency-rules.yaml` only  
**Forbidden files:** Any `.go` file; any other spec file  
**Definition of done:** 6 new entries for internal packages with `declared_callers`; `go run ./internal/checks/dependency-rules` passes  
**Required checks:** `go run ./internal/checks/dependency-rules`  
**Prompt for next agent:**
```
Add 6 new module entries to specs/dependency-rules.yaml:
- internal/semver: declared_callers: [x/ai/marketplace]
- internal/httputil: declared_callers: [middleware/internal/transport, x/gateway/cache, x/gateway/transform]
- internal/httpx: declared_callers: [x/gateway, x/messaging/webhook]
- internal/nethttp: declared_callers: [x/messaging]
- internal/stringsx: declared_callers: [x/messaging/webhook]
- internal/config: declared_callers: [x/observability/devtools]
Each entry: path, layer: internal, declared_callers list, deny: [reference/**, cmd/**].
After editing: go run ./internal/checks/dependency-rules — must pass with zero violations.
```

---

### PR 7: Add composable Makefile targets

**Title:** `build: add gates-go, check-boundaries, test-stable, bench Makefile targets`  
**Goal:** Let agents validate Go-only changes without running the website build.  
**Allowed files:** `Makefile` only  
**Forbidden files:** Any `.go` file; any YAML file  
**Definition of done:** `make check-boundaries`, `make test-stable`, `make gates-go` all pass without pnpm; `make gates` still runs full suite  
**Required checks:** `make check-boundaries` passes; `make test-stable` passes

---

### PR 8: `specs/public-api.yaml` initial draft

**Title:** `specs: add public-api.yaml — machine-readable stable GA surface`  
**Goal:** Authoritative list of stable exported symbols agents and checks reference.  
**Allowed files:** `specs/public-api.yaml` (create only)  
**Forbidden files:** Any `.go` file  
**Definition of done:** Valid YAML; entries for all 9 stable roots; symbols populated from corresponding `module.yaml:public_entrypoints`; `go run ./internal/checks/public-entrypoints-sync` passes  
**Required checks:** `go run ./internal/checks/public-entrypoints-sync`

---

### PR 9: Add `module.yaml` to internal packages

**Title:** `internal: add module.yaml to all internal packages`  
**Allowed files:** 7 new `module.yaml` files under `internal/*/`  
**Forbidden files:** Any `.go` file; any spec file  
**Definition of done:** 7 files created; each has `layer: internal`, `status: internal`, `declared_callers`; `go run ./internal/checks/module-manifests` passes  
**Required checks:** `go run ./internal/checks/module-manifests`

---

### PR 10: Make `contract.RouteStatePool` unexported

**Title:** `fix(contract): make RouteStatePool unexported — removes exported mutable global`  
**Goal:** Fix the one violation of the "no hidden globals" non-negotiable in a stable package.  
**Allowed files:** `contract/context_core.go`, `contract/freeze_test.go`, router file(s) that call `RouteStatePool`  
**Forbidden files:** Any spec file; any doc file; any file outside `contract/` or `router/`  
**Definition of done:** `rg 'RouteStatePool' . --include='*.go'` returns zero results outside comments; `go test -race ./contract/... ./router/...` passes; `freeze_test.go` updated  
**Required checks:**
```
go test -race -timeout 60s ./contract/... ./router/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/public-entrypoints-sync
```
**Prompt for next agent:**
```
Rename contract.RouteStatePool to unexported routeStatePool.
Steps:
1. rg -n 'RouteStatePool' . --include='*.go'  (list all sites)
2. In contract/context_core.go: rename var RouteStatePool → var routeStatePool
3. Add an unexported accessor in contract: func borrowRouteState() *RouteState { ... }
   and func returnRouteState(rs *RouteState) { ... } wrapping the unexported pool.
4. Update the router file that currently calls contract.RouteStatePool to call
   contract.borrowRouteState() / contract.returnRouteState() instead.
5. Update contract/freeze_test.go to remove RouteStatePool from expected exports.
6. go test -race -timeout 60s ./contract/... ./router/...  — must pass.
7. go run ./internal/checks/public-entrypoints-sync  — must pass.
```

---

## 10. Final Recommendation

The plumego repository is already in the top tier of agent-friendly Go repositories.
The specs, checks, module manifests, and recipes system is genuinely well-designed.
The work above is incremental refinement, not architectural overhaul.

**Priority order by agent-hours saved:**

1. **PRs 1–5 (agents/ directory)** — Agents currently read 5+ files before acting.
   These 5 files collapse that to 1–2. Highest leverage, lowest risk, docs-only.

2. **PR 7 (Makefile)** — Every single-module change requires an agent to know which
   subset of `make gates` to run. `make check-boundaries` removes that ambiguity.

3. **PR 6 (dependency-rules.yaml)** — Any agent running the dependency checker sees
   clean output even though 10 undeclared usages exist. This false negative will cause
   confusion the moment the checker is improved.

4. **PR 10 (RouteStatePool)** — The one concrete violation of the written rules in the
   stable layer. Leaves a credibility gap: agents reading the rules and finding this
   violation don't know whether the rules are enforced or aspirational.

5. **PRs 8–9** — Nice-to-have for completeness; lower urgency.

**What not to do:**
- Do not restructure the existing `specs/`, `tasks/`, or `docs/` hierarchy — it works.
- Do not merge `agents/` files into `AGENTS.md` — the separation between "rules for humans"
  and "distillation for agents" is the point.
- Do not move to a separate `tests/` directory — colocated tests are easier for agents
  to locate and modify in the same PR.
- Do not attempt all 10 PRs in one milestone — each is independently mergeable and
  must stay that way.
