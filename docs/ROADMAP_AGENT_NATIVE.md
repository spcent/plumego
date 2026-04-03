# Plumego Agent-Native Roadmap

Status: Draft  
Date: 2026-04-02  
Based on: [PRODUCT_PRD_AGENT_NATIVE.md](./PRODUCT_PRD_AGENT_NATIVE.md)

## 1. Purpose

This roadmap translates the agent-native product PRD into an execution plan.

It focuses on one question:

How do we turn Plumego from a repository with strong agent-aware architecture rules into a repository whose code layout, control plane, scaffolds, checks, and execution queue are all aligned enough that agents can iterate with low ambiguity and low rework?

This document complements the broader capability roadmap in `docs/ROADMAP.md`.
`docs/ROADMAP.md` tracks architectural and feature evolution across the repository.
This document tracks the product transformation of Plumego into an agent-native engineering system.

## 2. Planning Assumptions

This roadmap assumes:

- architecture correctness is more important than compatibility preservation
- the canonical path should be simplified even if some existing layouts must be cleaned up
- agent productivity is a first-class product metric
- machine-readable rules should be preferred over prose wherever possible
- each implementation slice should be reversible and small enough for one focused agent pass

## 3. Current State as of 2026-04-02

Plumego already has strong foundations:

- stable roots with explicit boundaries
- `specs/*` as a machine-readable control plane
- `module.yaml` across major modules
- a canonical reference app
- scaffold and generation tooling
- internal architecture checks
- significant migration of mixed concerns from stable roots into `x/*`

However, the repository is not yet fully agent-native because the following gaps remain:

- the control plane does not fully reflect the current code topology
- some extension capabilities exist in code but are not fully registered in specs and primers
- package-level responsibility is still weaker than module-level responsibility
- current checks are stronger on import boundaries than on semantic responsibility boundaries
- task execution is not yet driven by a live active card queue
- product metrics for agent productivity and rework are not yet part of the operating model

## 4. Roadmap Outcome

At the end of this roadmap, Plumego should behave like this:

- an agent can classify a task and immediately find the correct package-level starting point
- stable roots contain only durable primitives and no app-facing capability leakage
- every extension family has a complete canonical discovery path in docs, specs, and manifests
- active roadmap work is represented as task cards, not only prose
- checks fail when implementation drifts from architectural intent
- scaffold, reference, and generated output stay aligned by construction
- maintainers can measure whether agent work is becoming faster and less error-prone

## 5. Workstreams

This roadmap is organized into five workstreams:

1. Control-plane truth alignment
2. Package-level control plane
3. Responsibility enforcement and drift prevention
4. Task execution operating system
5. Agent productivity instrumentation

The phases below sequence these workstreams in the order that most reduces ambiguity early.

## 6. Phase Overview

| Phase | Name | Primary Outcome |
| --- | --- | --- |
| Phase 1 | Align Control Plane With Current Code | Specs, docs, manifests, and architecture docs describe the current capability topology truthfully |
| Phase 2 | Add Package-Level Control Plane | Agents can choose the correct package, not only the correct module |
| Phase 3 | Enforce Responsibility Boundaries | Checks block semantic drift, not only import drift |
| Phase 4 | Establish Task-Card Operating Model | Roadmap work becomes executable through active cards with lifecycle rules |
| Phase 5 | Converge Reference, Scaffold, and Generation | Canonical structure becomes reproducible by default |
| Phase 6 | Instrument and Gate Agent-Native Readiness | Product metrics and release gates measure real agent productivity and rework |

## 7. Detailed Phases

### Phase 1: Align Control Plane With Current Code

Status: next

#### Objective

Bring `docs/`, `specs/`, manifests, and taxonomy metadata into agreement with the current repository shape so that the control plane stops lagging behind implementation.

#### Why this phase is first

The current repository already moved several concerns into better locations, but the machine-readable and human-readable control surfaces have not completely caught up. If this is not fixed first, later package-level work will be built on stale topology data.

#### Known gaps this phase should close

- `x/fileapi` exists in code but is not yet fully represented in the extension taxonomy
- `x/fileapi` does not yet have a matching module primer under `docs/modules/`
- the architecture blueprint still reflects an older extension list
- `contract/protocol` remains as an empty residue directory and is inconsistent with the blueprint's own rules against broad protocol-family roots

#### Deliverables

- register `x/fileapi` in the extension layer metadata
- decide and document whether `x/fileapi` is:
  - a standalone extension root, or
  - a sub-capability of `x/data`
- add `docs/modules/x-fileapi/README.md`
- update `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md` to reflect the actual extension topology
- update `specs/repo.yaml` and `specs/extension-entrypoints.yaml` accordingly
- update `specs/agent-entrypoints.yaml` with file API discovery rules if needed
- remove or explicitly forbid the empty `contract/protocol` residue
- add a check that flags orphaned extension roots and empty misleading package directories

#### Dependencies

- none

#### Exit Criteria

- every shipping extension root appears in `specs/repo.yaml`
- every app-facing extension root appears in the appropriate taxonomy metadata
- every extension root has either a primer or an explicit rationale for not having one
- no empty misleading package directory remains in the canonical tree

#### Non-goals

- do not redesign capability boundaries that are already correct in code
- do not introduce package-level manifests in this phase

### Phase 2: Add Package-Level Control Plane

Status: planned

#### Objective

Extend Plumego's machine-readable control plane from module-level guidance to package-level guidance for high-risk or high-traffic areas.

#### Why this phase matters

Agents currently know how to choose the right module most of the time. They still need too much local interpretation to choose the right package inside a complex module or extension family.

#### Deliverables

- introduce a package-level metadata format for selected hot paths
- establish `specs/package-index.yaml` as the package-level hotspot discovery surface
- start with hot packages where wrong entrypoint cost is high
- define for each package:
  - layer
  - role
  - responsibilities
  - non-goals
  - allowed concerns
  - forbidden concerns
  - canonical start files
  - avoid paths
  - default validations
- add a repository-level package index spec or equivalent discoverability mechanism
- connect package metadata to task-family entrypoints where ambiguity remains high

#### Suggested first package groups

- `x/fileapi`
- `x/data/file`
- `x/rest`
- `x/gateway`
- `x/tenant/core`
- `x/tenant/store/db`
- `middleware/httpmetrics`
- `middleware/requestid`
- `contract`

#### Dependencies

- Phase 1 control-plane truth alignment

#### Exit Criteria

- agents can classify common tasks to the correct package without broad code search
- at least the top ambiguity hotspots have package-level metadata
- package-level metadata is discoverable from one repo-native control surface

#### Non-goals

- do not add package manifests for every leaf package immediately
- do not turn package metadata into a second prose documentation system

### Phase 3: Enforce Responsibility Boundaries

Status: in progress

#### Objective

Upgrade repo checks so they catch semantic responsibility drift, not only import drift.

#### Why this phase matters

Import rules are necessary but insufficient. A package can still violate architecture while importing only allowed modules.

#### Deliverables

- add checks for empty misleading package directories
- add checks for orphaned extension roots not registered in the control plane
- add checks for app-facing HTTP surface leakage into stable roots
- add checks for protocol-family catch-all leakage into stable roots
- add checks for module primer and package metadata coverage for registered app-facing roots
- add checks for task-family coverage where new app-facing roots are introduced

#### Progress So Far

- empty misleading package directory checks are in place
- orphaned extension root checks are in place
- module primer coverage for canonical app-facing extension roots is now enforced
- `specs/package-index.yaml` package path and `start_with` path coverage is now enforced
- stable-root app-facing HTTP surface leakage checks are now in place
- task-family coverage checks remain open

#### Candidate semantic rules

- stable `store/*` must not expose app-facing HTTP handlers
- stable `contract/*` must not grow protocol-family namespaces
- stable roots must not contain route registration helpers for extension capabilities
- app-facing extension roots must be declared in taxonomy metadata
- new extension roots must ship with module manifest plus primer

#### Dependencies

- Phase 1 for topology truth
- Phase 2 for package-level metadata where needed

#### Exit Criteria

- major architecture mistakes fail in checks before review
- reviewers no longer need to rely on manual architectural rediscovery for common drift types

#### Non-goals

- do not attempt to encode every architectural nuance into one checker pass
- do not overfit checks to temporary implementation details

### Phase 4: Establish Task-Card Operating Model

Status: in progress

#### Objective

Turn roadmap execution into a live card-driven system so agents operate from active execution units rather than from broad prose documents.

#### Why this phase matters

Plumego already has card format and completed-card archives. It does not yet operate with an active, continuously maintained card queue as the default execution surface.

#### Deliverables

- define `tasks/cards/active/` as the working queue
- define lifecycle states for cards:
  - active
  - blocked
  - done
  - superseded
- add checks for card format and discoverability
- connect roadmap phases to active cards
- define cardinality rules:
  - one primary module
  - up to five files
  - up to three validation commands
- define card status ownership and archival rules

Completed so far:

- `tasks/cards/active/` now exists as the live queue
- near-term agent-native work is represented by active cards instead of only roadmap prose
- base queue documentation now distinguishes the live queue from the completed archive

Remaining work:

- add queue-specific checks for card format and discoverability
- decide whether `blocked/` and `superseded/` need dedicated physical directories or remain logical states
- tighten queue ordering rules so the next execution card remains obvious as the queue grows

#### Dependencies

- Phase 1 to ensure topology truth

#### Exit Criteria

- all near-term roadmap work exists as active or blocked cards
- agents can move from roadmap to an executable card without extra decomposition
- completed work archives cleanly without leaving stale active items

#### Non-goals

- do not turn cards into long design documents
- do not mix large multi-module programs into one card

### Phase 5: Converge Reference, Scaffold, and Generation

Status: planned

#### Objective

Make the canonical structure reproducible by default across reference code, scaffold output, and generated code.

#### Why this phase matters

Reference and scaffold are already present, but the agent-native product is not complete until code generation and scaffolding also participate in the same control plane and same drift guarantees.

#### Deliverables

- define structure-level invariants for canonical scaffold output
- add drift checks between reference app and canonical scaffold
- define when generators may emit package-local helpers versus app-local wiring
- ensure generated output never creates non-canonical handler, route, or middleware styles
- add at least one generator contract test that asserts canonical route registration shape

#### Dependencies

- Phase 2 package-level control plane
- Phase 3 enforcement checks

#### Exit Criteria

- generated services start in the canonical structure with no cleanup pass required
- scaffold output and reference structure stay aligned by checks
- generators do not introduce new architecture styles accidentally

#### Non-goals

- do not expand feature breadth just to exercise generators
- do not create multiple competing scaffold styles

### Phase 6: Instrument and Gate Agent-Native Readiness

Status: planned

#### Objective

Add measurable product metrics and a release gate so agent-native readiness can be evaluated empirically rather than by intuition.

#### Why this phase matters

Without product metrics, the repository can look cleaner while still failing to reduce rework or ambiguity in practice.

#### Deliverables

- define a small first set of operational metrics
- instrument or derive the following where feasible:
  - average files touched per routine task
  - percentage of tasks completed within one primary module
  - percentage of roadmap work represented as active cards
  - architecture drift findings caught by checks
  - scaffold cleanup required after generation
- define a lightweight agent-native readiness checklist
- add a release review step for agent-native readiness

#### Dependencies

- all prior phases

#### Exit Criteria

- maintainers can describe agent-native progress with data
- release readiness includes structural and workflow metrics, not only code correctness

#### Non-goals

- do not build heavy telemetry infrastructure inside the repo
- do not block iteration on metrics perfection

## 8. Milestone Sequencing

### Milestone A: Control-Plane Truth

Includes:

- Phase 1 complete

Meaning:

- the repository stops telling two different stories about where capabilities live

### Milestone B: Package Discovery Determinism

Includes:

- Phase 2 complete
- core parts of Phase 3 complete

Meaning:

- agents can choose the correct package, not only the correct module

### Milestone C: Executable Operating Model

Includes:

- Phase 4 complete
- core parts of Phase 5 complete

Meaning:

- roadmap, cards, scaffold, reference, and generation form one closed loop

### Milestone D: Measured Agent-Native Readiness

Includes:

- Phase 6 complete

Meaning:

- Plumego can claim agent-native readiness based on measurable operational evidence

## 9. Prioritization Rules

When tradeoffs appear, prioritize in this order:

1. remove ambiguity before adding new capability breadth
2. update control plane before adding more implementation surface
3. enforce boundaries before documenting exceptions
4. create executable cards before expanding roadmap prose
5. instrument key metrics before optimizing secondary workflows

## 10. Risks and Mitigations

### Risk 1: Specs lag implementation again

Mitigation:

- phase-gate new app-facing roots behind spec and primer registration checks

### Risk 2: Package metadata grows too large

Mitigation:

- start only with ambiguity hotspots
- keep metadata machine-readable and minimal

### Risk 3: Task cards become stale ceremony

Mitigation:

- add active queue lifecycle rules
- tie cards directly to roadmap milestones and review expectations

### Risk 4: Checkers become brittle

Mitigation:

- add rules incrementally
- focus on high-signal semantic violations first

## 11. Initial Execution Slice

The first execution slice should focus on control-plane truth alignment, because later phases depend on it.

### Now

#### Card AN-0101

Goal:
- Align the control plane with the existing file capability split and remove the remaining empty protocol residue.

Scope:
- extension taxonomy for `x/fileapi`
- architecture and repo metadata alignment
- cleanup rule for empty misleading directories

Non-goals:
- Do not redesign the current file capability split.
- Do not add package-level manifests yet.

Files:
- `specs/repo.yaml`
- `specs/extension-entrypoints.yaml`
- `specs/agent-entrypoints.yaml`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `docs/modules/x-fileapi/README.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- Keep the blueprint, extension taxonomy, and new primer aligned on where file API work starts.

Done Definition:
- `x/fileapi` is a first-class declared capability in the control plane.
- Agents no longer have to infer file capability entrypoints from code alone.

### Queue

#### Card AN-0102

Goal:
- Define the first package-level metadata spec for ambiguity hotspots.

Scope:
- package-level metadata format
- discoverability rules
- first hotspot package list

Non-goals:
- Do not annotate the entire repo.

Files:
- `specs/package-index.yaml`
- `specs/agent-entrypoints.yaml`
- `docs/PRODUCT_PRD_AGENT_NATIVE.md`
- `docs/ROADMAP_AGENT_NATIVE.md`

Tests:
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Keep package-level control plane intent aligned between PRD and roadmap.

Done Definition:
- Plumego has a defined package-level metadata contract for hotspot packages.
- The repository exposes one discoverable place to find package-level entrypoint guidance.
- The first hotspot package list is explicit and bounded.

Initial hotspot target list:

- `x/fileapi`
- `x/data/file`
- `x/rest`
- `x/gateway`
- `x/tenant/core`
- `x/tenant/store/db`
- `middleware/httpmetrics`
- `middleware/requestid`
- `contract`

#### Card AN-0103

Goal:
- Add a checker for orphaned extension roots and empty misleading directories.

Scope:
- orphaned extension root detection
- empty misleading directory detection
- failure messaging

Non-goals:
- Do not add package semantic rules yet.

Files:
- `internal/checks/agent-workflow/main.go`
- `internal/checks/checkutil/checkutil.go`
- `internal/checks/checkutil/checkutil_test.go`
- `specs/repo.yaml`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Keep checker behavior aligned with control-plane rules.

Done Definition:
- New extension roots and misleading empty dirs cannot silently drift into the repo.

#### Card AN-0104

Goal:
- Establish the active task-card queue and lifecycle rules.

Scope:
- active queue layout
- state model
- archival rules

Non-goals:
- Do not migrate the full roadmap in one card.

Files:
- `tasks/cards/README.md`
- `tasks/cards/active/README.md`
- `docs/ROADMAP_AGENT_NATIVE.md`

Tests:
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Keep active queue rules aligned with roadmap execution expectations.

Done Definition:
- Plumego has a documented live execution surface under `tasks/cards/active/`.
- Card lifecycle expectations are clear enough for agents and maintainers to use consistently.
- New near-term roadmap work can be added to the active queue without ad hoc conventions.

Outcome:
- `tasks/cards/README.md` defines the lifecycle vocabulary for `active`, `blocked`, `done`, and `superseded`.
- `tasks/cards/active/README.md` defines how the live queue should be maintained while only `active/` and `done/` exist physically.
- Phase 4 is now in progress rather than purely planned because the live queue and its operating rules now exist.

## 12. Agent-Native Readiness Definition

Plumego should only claim agent-native readiness after all of the following are true:

- the control plane matches the actual code topology
- package-level discovery exists for ambiguity hotspots
- semantic responsibility checks block common drift
- task cards are the default execution surface
- scaffold, reference, and generation are structurally aligned
- agent productivity and rework metrics are visible in release reviews

Until then, Plumego should describe itself as progressing toward agent-native readiness rather than complete.
