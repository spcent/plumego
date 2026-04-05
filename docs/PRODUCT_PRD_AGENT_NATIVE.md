# Plumego Product PRD

Status: Draft  
Date: 2026-03-17  
Owner: Plumego Maintainers  
Audience: Product, Architecture, Runtime, Security, DX, AI-Agent Workflow

## 1. Product Summary

Plumego is an agent-native Go web toolkit built on the standard library HTTP model.

Its product goal is not only to provide runtime capabilities such as routing, middleware, security, storage primitives, and optional `x/*` extensions, but to make the entire repository itself operable by AI agents with low ambiguity, low regression risk, and low rework cost.

In Plumego's target state, an agent should be able to:

- identify the correct entrypoint for a task without broad repo search
- understand package responsibility without relying on tribal knowledge
- make small reversible changes in one clear implementation path
- run the correct validations automatically
- avoid cross-module leakage and avoid architectural backtracking

This PRD defines Plumego as a product for agent-driven software development, not only as a Go library.

## 2. Vision

Build the most agent-friendly Go web toolkit and repository operating model, where:

- the stable core remains explicit, small, and durable
- app-facing capabilities are discoverable through one canonical path
- agents can iterate on product features quickly without structural guesswork
- architecture drift is blocked mechanically rather than corrected socially
- new services and extensions inherit the same explicit structure by default

## 3. Product Positioning

### 3.1 Product Category

Plumego is:

- a standard-library-first Go HTTP toolkit
- an agent-native repository and workflow system
- a reference architecture and scaffolding baseline for explicit service construction

Plumego is not:

- a hidden-magic full-stack framework
- a plugin-first runtime with implicit registration
- a feature catalog where all product concerns are mixed into stable roots

### 3.2 Core Value Proposition

For human engineers:

- explicit control flow
- predictable package ownership
- low-regret refactoring
- reusable reference and scaffold system

For AI agents:

- deterministic discovery
- machine-readable boundaries
- single canonical implementation path
- minimal search and edit radius
- test and validation determinism

## 4. Background and Problem Statement

Traditional framework repositories are difficult for agents because they often have:

- multiple equally plausible entrypoints
- mixed responsibilities inside shared packages
- implicit registration and runtime magic
- documentation that disagrees with actual code placement
- broad helper directories that become dumping grounds
- unclear test ownership and validation order

These issues lead to repeated costs:

- agents start from the wrong module
- changes span too many files
- stable boundaries erode over time
- similar features are implemented in inconsistent ways
- follow-up refactors are required to restore order
- roadmap items are delayed by repeated rediscovery

Plumego already addresses part of this through `docs/`, `specs/`, `module.yaml`, `reference/standard-service`, and internal checks. However, the product has not yet reached a state where repository structure, execution model, scaffolds, and checks form one closed loop.

The product problem is therefore:

How do we turn Plumego from an agent-aware toolkit into a fully agent-native product where the repository itself is optimized for AI-driven implementation, validation, and iteration?

## 5. Product Goals

### 5.1 Primary Goal

Make Plumego the default architecture and execution substrate for agent-driven Go service development.

### 5.2 Strategic Goals

- make repository discovery deterministic
- make boundaries machine-enforceable
- reduce implementation ambiguity to one canonical path per task family
- reduce average change radius and rollback cost
- make reference, scaffold, docs, and checks converge on the same architecture
- enable roadmap execution through small, reversible, agent-safe work units

### 5.3 Success Definition

Plumego is successful when an agent can receive a feature request and, with minimal extra instruction:

- choose the correct start path
- touch the correct module only
- preserve architecture without human correction
- pass the required checks
- produce no structural cleanup follow-up work

## 6. Target Users

### 6.1 Primary Users

- maintainers of Plumego itself
- engineers building services on Plumego
- AI coding agents operating on Plumego repositories

### 6.2 Secondary Users

- architecture reviewers
- platform and runtime teams
- template and scaffold authors
- documentation and DX maintainers

### 6.3 User Personas

#### Persona A: Repo Maintainer

Needs:

- architecture stability
- clear ownership
- controlled extension growth
- mechanical enforcement of boundaries

Pain today:

- new capabilities can still land in structurally wrong places
- docs and implementation may drift
- migration debt can remain visible as if it were canonical

#### Persona B: Product Engineer

Needs:

- a clear way to add endpoints, middleware, tenant logic, messaging, gateway, or resource APIs
- a reference layout that is easy to copy
- reliable test and validation guidance

Pain today:

- capability discovery can still require reading multiple docs
- stable roots can still contain mixed signals

#### Persona C: AI Agent

Needs:

- one obvious read path
- one obvious write path
- machine-readable task entrypoints
- explicit non-goals and forbidden zones
- deterministic validation sequence

Pain today:

- some packages still carry mixed transport, tenant, and storage concerns
- some directories remain as misleading historical or migration artifacts
- not every product task has an active execution card

## 7. Product Principles

The following are non-negotiable product principles:

- stdlib first
- explicit over implicit
- one obvious way
- one canonical app path
- stable roots stay narrow
- capability families expose one canonical entrypoint
- task execution should be card-driven and reversible
- repo rules must be machine-readable before they are prose-heavy
- architecture drift must be blocked by checks, not only by review comments

## 8. Product Scope

### 8.1 In Scope

- stable root packages as durable public primitives
- `x/*` extension families as app-facing capability entrypoints
- canonical reference applications
- scaffold and code generation system aligned with canonical structure
- machine-readable specs for ownership, dependency rules, task entrypoints, and change recipes
- machine-readable hotspot package index for package-level discovery
- repo-native task cards for reversible execution
- architecture checks and validation gates
- documentation that mirrors the actual canonical path

### 8.2 Out of Scope

- supporting multiple equal bootstrap styles
- introducing hidden runtime registration or `init()` discovery
- treating experimental extensions as part of the stable kernel
- preserving legacy mixed-boundary layouts only for convenience
- allowing stable roots to become feature catalogs
- using prose-only governance where a machine-readable contract can exist

## 9. Product Requirements

### 9.1 Repository Information Architecture

The repository must expose a deterministic information architecture:

- `docs/` for human explanation
- `specs/` for machine-readable contracts
- `tasks/` for execution work units
- `reference/` for canonical and explicitly non-canonical examples
- stable library roots at top level
- extension capability roots under `x/*`

Requirements:

- every top-level control surface must have a singular purpose
- no top-level legacy buckets may compete with stable roots or `x/*`
- every new module must declare ownership, allowed imports, doc paths, and validations
- every capability family must define its canonical discovery entrypoint

### 9.2 Stable Root Product Requirements

Stable roots must satisfy all of the following:

- narrow responsibility
- minimal public surface
- no dependency on `x/*`
- no app-facing feature orchestration
- no hidden service locator or implicit injection
- no tenant-aware policy behavior in stable `middleware` or `store`
- no protocol-family gateway concerns in `contract`
- no HTTP endpoint ownership in `health`

### 9.3 Extension Product Requirements

Each extension family must:

- have one canonical family entrypoint
- clearly distinguish app-facing entrypoints from subordinate primitives
- define its own module manifest and docs
- avoid redefining canonical bootstrap
- remain optional relative to the stable kernel

### 9.4 Reference and Scaffold Requirements

Plumego must provide:

- one canonical reference app
- non-canonical feature references outside the canonical path
- scaffold templates generated from the same design rules as the canonical reference
- drift checks ensuring scaffolds and canonical reference do not diverge structurally

### 9.5 Agent Workflow Requirements

An agent must be able to answer the following without broad repo exploration:

1. What type of task is this?
2. Which module or package should I start with?
3. Which places should I avoid?
4. What is the canonical implementation shape?
5. What validations must run?
6. What docs and manifests must stay in sync?

This requires:

- task-family entrypoint metadata
- hotspot package index metadata
- ownership metadata
- change recipes
- active task cards
- package and module manifests

The first package-level discovery surface should live in:

- `specs/package-hotspots.yaml`

This index should stay intentionally sparse and cover only packages where the wrong first package choice causes meaningful rework.

### 9.6 Task Execution Requirements

All non-trivial roadmap work should be executable through task cards.

Task card requirements:

- one primary module whenever possible
- small file set
- clear non-goals
- explicit test list
- explicit done definition
- reversible in one focused implementation pass

### 9.7 Validation Requirements

Validation must operate at multiple levels:

- package and module tests
- dependency and boundary checks
- manifest and docs sync checks
- reference and scaffold drift checks
- quality gates for repo-wide confidence

Validation must detect:

- import boundary violations
- canonical path drift
- package responsibility drift
- undocumented public surface growth
- stale or misleading task and manifest metadata

## 10. Non-Functional Requirements

### 10.1 Determinism

Two competent agents given the same task should choose substantially the same start path, file set, and validation order.

### 10.2 Low Search Radius

Routine feature work should require reading only:

- canonical style guide
- relevant task-family entrypoint
- target module or package manifest
- local code in the owning area

### 10.3 Reversibility

Most changes should be decomposable into small cards that can be reverted without broad repo fallout.

### 10.4 Traceability

Every architectural decision should map cleanly to:

- a manifest
- a spec
- a check
- a reference or scaffold example

### 10.5 Security

Product behavior must preserve current security rules:

- fail closed on auth, verification, and policy errors
- never log secrets
- avoid service-locator patterns
- preserve timing-safe secret checks

### 10.6 Performance of the Workflow

Repo checks must remain fast enough for frequent agent iteration.

The product should optimize for:

- short feedback loops
- selective validation first
- full gate execution when needed

## 11. Functional Capability Model

### 11.1 Stable Kernel Capabilities

Plumego stable kernel includes:

- app lifecycle
- routing
- transport contracts
- narrow transport middleware
- security helpers
- storage primitives
- health/readiness models
- logging and metrics contracts

These are the long-lived foundation and must remain boring, explicit, and durable.

### 11.2 Extension Capability Families

Extension families include:

- tenant
- messaging
- gateway
- resource API standardization
- frontend
- observability
- ops/admin surfaces
- discovery
- websocket
- AI and other evolving capabilities

These may evolve quickly, but each family must still preserve one canonical entrypoint.

### 11.3 Product Workflow Capabilities

Plumego as a product must support:

- repo discovery
- task classification
- canonical code generation
- reference consumption
- scaffold generation
- mechanical architectural validation
- roadmap-to-card execution

## 12. User Journeys

### 12.1 Journey A: Add an HTTP Endpoint

Expected flow:

- agent reads the style guide
- agent reads HTTP endpoint entrypoint metadata
- agent reads the app-local route wiring and owning module manifest
- agent implements handler in the canonical transport shape
- agent runs targeted validation first
- agent syncs docs only if public behavior changed

Success criteria:

- route registration is explicit
- no business logic lands in middleware
- no new alternative endpoint style is introduced

### 12.2 Journey B: Add a New Capability Family

Expected flow:

- maintainer creates new `x/*` family only if it does not fit an existing family
- manifest, docs, and entrypoint metadata are added first
- reference and scaffold impact are defined
- task cards are created for incremental landing

Success criteria:

- no competition with existing family entrypoints
- no stable-root pollution
- agent can find the new family from `specs/extension-taxonomy.yaml`

### 12.3 Journey C: Generate a New Service

Expected flow:

- engineer or agent starts from scaffold
- scaffold mirrors canonical reference
- generated app uses explicit wiring and standard module boundaries
- no hidden compatibility or legacy mode is introduced

Success criteria:

- generated code matches canonical structure
- no follow-up cleanup is required before adding product work

### 12.4 Journey D: Execute Roadmap Work

Expected flow:

- roadmap item is already split into task cards
- agent chooses active card
- card points to the correct module and validations
- implementation is reversible and independently testable

Success criteria:

- roadmap execution does not require rediscovery
- cards reflect the real canonical implementation path

## 13. Key Product Decisions

Plumego will choose:

- one canonical app structure instead of many examples with equal weight
- one canonical entrypoint per capability family instead of sibling-package discovery ambiguity
- machine-readable specs instead of prose-only governance
- narrow stable roots instead of convenience catalogs
- explicit route and dependency wiring instead of hidden framework magic
- task cards as execution units instead of large, vague roadmap-only planning

## 14. Product Risks

### 14.1 Architecture Drift Risk

Risk:

- code placement drifts away from manifests and specs

Impact:

- agent decisions become inconsistent
- roadmap velocity drops

Mitigation:

- stronger checks
- package-level manifests
- scaffold and reference drift enforcement

### 14.2 Mixed-Boundary Package Risk

Risk:

- stable packages absorb transport, tenant, provider, or feature orchestration concerns

Impact:

- search radius expands
- rework increases

Mitigation:

- move app-facing surfaces into `x/*`
- add responsibility checks beyond import rules

### 14.3 Documentation Drift Risk

Risk:

- README, module primers, reference code, and specs disagree

Impact:

- agents follow the wrong source of truth

Mitigation:

- doc-path checks
- canonical-source policy
- automated sync checks

### 14.4 Capability Family Proliferation Risk

Risk:

- multiple sibling packages compete as entrypoints

Impact:

- discovery ambiguity

Mitigation:

- explicit family-level canonical entrypoints
- deprecation and convergence policy

## 15. Success Metrics

### 15.1 Product Metrics

- percentage of tasks that can be completed within one primary module
- percentage of roadmap items represented by active task cards
- percentage of feature work started from the correct canonical entrypoint on first attempt
- average files touched per routine feature task
- number of architecture-drift findings caught by checks before review
- number of follow-up cleanup PRs caused by wrong placement

### 15.2 Agent Workflow Metrics

- median reads before first code change
- median targeted validations before full-gate run
- percentage of tasks with deterministic validation sequence
- percentage of generated services that require no structural cleanup

### 15.3 Quality Metrics

- stable-root boundary violations per release
- docs/spec/reference drift findings per release
- scaffold/reference structural drift findings
- public API surface changes without manifest/doc updates

## 16. Release Criteria for Agent-Native v1+

Plumego can be considered agent-native at product level when all of the following are true:

- stable roots are free of known mixed-boundary packages
- every capability family has one canonical discovery path
- every active roadmap item exists as task cards
- scaffold output matches the canonical reference structurally
- checks enforce both dependency and responsibility boundaries
- docs, manifests, reference, and specs agree on the same canonical path
- routine feature work can be completed by agents with low variance

## 17. Explicit Non-Goals

Plumego does not aim to:

- optimize for hidden convenience over explicitness
- support many parallel framework styles equally
- embed all fast-moving capabilities into the stable kernel
- preserve structurally wrong package placement for backward comfort
- rely on review comments as the primary architecture enforcement mechanism

## 18. Dependencies and Constraints

### 18.1 Technical Constraints

- preserve `net/http` compatibility
- stable main module remains dependency-light and stdlib-first
- explicit DI and route wiring remain default
- checks must be automatable inside the repo

### 18.2 Organizational Constraints

- maintainers must be able to explain and defend each boundary
- docs must remain concise enough for agents to consume
- roadmap work must be incremental and reversible

## 19. Open Questions for Roadmap Planning

The following questions should be resolved in the roadmap phase:

- should package-level manifests be introduced for all non-trivial subpackages or only selected hot spots?
- what is the exact product boundary between stable `store` primitives and future `x/data` or similar capability roots?
- should scaffold generation be fully declarative from manifests and specs, or remain template-first with checks?
- what metrics should be automated first to quantify agent productivity and rework reduction?
- what remaining mixed-boundary packages must be restructured before claiming agent-native readiness?

## 20. Roadmap Planning Implications

The roadmap derived from this PRD should prioritize work in this order:

1. eliminate structural ambiguity
2. remove mixed-boundary packages from stable roots
3. strengthen checks from import boundaries to responsibility boundaries
4. complete task-card-driven execution coverage
5. harden scaffold, reference, and generation convergence
6. instrument product metrics for agent productivity and rework reduction

The roadmap should not begin with broad feature expansion. It should begin with reducing ambiguity and locking the operating model.

## 21. Appendix: Product Statement

Plumego is a standard-library-first, explicit, agent-native Go toolkit and repository system designed so that both humans and AI agents can build, evolve, and validate services through one canonical path with minimal ambiguity and minimal rework.
