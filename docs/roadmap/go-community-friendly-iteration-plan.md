# Go Community-Friendly Iteration Plan

**Version:** 1.0  
**Status:** Active  
**Last Updated:** June 2026

This document outlines Plumego's strategic iteration plan to position the project as the standard-library-first HTTP toolkit for Go developers who value explicit, maintainable, agent-friendly services.

---

## Executive Summary

Plumego is a production-ready HTTP toolkit that provides more structure than raw `net/http` without the opinions and overhead of larger frameworks. This plan focuses on **community accessibility**, **clear positioning**, and **sustainable growth** through documentation, governance clarity, and evidence-based decision-making.

### Key Objectives

1. **Clarity:** Make it obvious what Plumego is, who it's for, and why it exists.
2. **Stability:** Provide clear compatibility guarantees so developers can adopt with confidence.
3. **Community:** Build pathways for contribution and extension without compromising core values.
4. **Evidence:** Record decisions and rationale in machine-checkable specifications.

---

## Community Value Assessment

### Current Strengths

1. **Standard Library First:** Handlers are ordinary `http.Handler`, middleware follow `func(http.Handler) http.Handler`, no framework-specific coupling.
2. **Explicit Wiring:** Routes, middleware, and dependencies are visible at construction; no hidden magic or auto-discovery.
3. **Small Stable Core:** Only 9 stable root packages; optional `x/*` extensions for capabilities.
4. **Mature Codebase:** Tested, benchmarked, documented; production-ready.
5. **Agent-Friendly Control Plane:** Machine-checkable boundaries, task specs, automated validation.

### Market Position

**Plumego fills the gap between:**
- Raw `net/http` (too low-level; requires manual routing, middleware composition, error handling)
- Large frameworks like Echo, Gin, Fiber (too opinionated; require learning framework idioms; vendor lock-in risk)

**Target Developer:**
- Go teams building JSON APIs, backend services, or microservices.
- Teams that value stdlib compatibility and explicit application wiring.
- Teams building services for long-term maintenance with minimal churn.
- DevOps-oriented projects (agent-friendly, CI/CD-native).

### Addressable Market

1. **Short-term (v1.x):** Teams building services from scratch who want stdlib + structure (estimated 10-20% of new Go projects).
2. **Medium-term (v2.x):** Teams migrating from larger frameworks (estimated 5-10% of existing projects).
3. **Long-term:** Plumego becomes a de facto standard for "how to build structured Go services."

---

## Positioning

### Messaging

**Elevator Pitch:**
"Plumego is the standard-library-first HTTP toolkit for Go. Build explicit, maintainable services without framework overhead or vendor lock-in."

**Detailed Positioning:**
1. **Built on stdlib:** Handlers are `http.Handler`, middleware follow `func(http.Handler) http.Handler`. No special types, no coupling.
2. **Small stable core:** Start with 9 packages; add optional `x/*` extensions only when you need them.
3. **Explicit wiring:** Routes, middleware, and dependencies are visible in your code. No hidden magic or auto-discovery.
4. **Production-ready:** Full test coverage, benchmarked against Chi and Gin, compatible with existing middleware and tools.
5. **Agent-friendly maintenance:** Machine-checkable boundaries, task specs, and automated validation for AI-assisted development.

### Key Differentiators

| Aspect | Plumego | Gin/Echo/Fiber | Chi |
|---|---|---|---|
| Handler type | `http.Handler` | Framework-specific | `http.Handler` |
| Middleware type | `func(http.Handler) http.Handler` | Framework-specific | `func(http.Handler) http.Handler` |
| Stable roots | 9 packages (GA) | Monolithic | 1 package (core) |
| Extensions | Optional `x/*` | Built-in, all or nothing | Ad-hoc |
| Explicit wiring | Yes, always | Implicit defaults | Yes |
| Stdlib compatibility | 100% (built on it) | Requires wrapper | High |
| Lock-in risk | Low (stdlib components) | High (framework types) | Low |

---

## Roadmap and Iteration Plan

### Phase 1: Documentation & Governance (v1.1.0 → v1.2.0)

**Duration:** Q2–Q3 2026  
**Focus:** Clarity and community onboarding

#### Deliverables

1. **Documentation Overhaul**
   - [ ] `README.md` — Emphasize stdlib-first positioning
   - [ ] `STABILITY.md` — Full compatibility policy
   - [ ] `COMPATIBILITY.md` — How Plumego integrates with stdlib and ecosystem
   - [ ] `docs/ADOPTION_PATH.md` — Progressive learning path
   - [ ] `docs/EXTENSION_MATURITY.md` — Human-readable maturity dashboard

2. **Examples**
   - [ ] `examples/hello/` — Minimal "hello world" with no dependencies
   - [ ] `examples/json-api/` — Small JSON CRUD service
   - [ ] `examples/with-middleware/` — Custom middleware example
   - [ ] Examples compile and run without setup

3. **Governance Files**
   - [ ] Update `AGENTS.md` with community contribution pathways
   - [ ] Define extension authoring contract in `docs/guides/extension-authoring.md`
   - [ ] Create `docs/CONTRIBUTING.md` — Contribution guidelines

4. **Migration Guides**
   - [ ] Finalize `docs/guides/migration/from-chi.md`
   - [ ] Finalize `docs/guides/migration/from-gin.md`
   - [ ] Finalize `docs/guides/migration/from-echo.md`
   - [ ] Finalize `docs/guides/migration/middleware-compat.md`

#### Validation

- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes with no warnings
- [ ] All examples compile and run
- [ ] README review by 2–3 external Go developers
- [ ] Extension maturity dashboard in sync with `module.yaml` manifests

---

### Phase 2: Community Engagement (v1.2.0 → v1.3.0)

**Duration:** Q3–Q4 2026  
**Focus:** Adoption pathways and feedback loops

#### Deliverables

1. **Community Programs**
   - [ ] Announce Plumego in Go communities (r/golang, Go Forum, Go Slack)
   - [ ] Publish adoption story or case study
   - [ ] Host optional office hours or Q&A session

2. **Adoption Pathways**
   - [ ] Create "migrate from Chi" checklist
   - [ ] Create "migrate from Gin" checklist
   - [ ] Create "build your first service" walkthrough
   - [ ] Publish decision tree: "Is Plumego right for me?"

3. **Feedback Collection**
   - [ ] GitHub issue template for "documentation feedback"
   - [ ] Survey: "What would make Plumego easier to adopt?"
   - [ ] Metrics: GitHub stars, issues, PRs, fork trends

#### Validation

- [ ] Minimum 20 GitHub stars
- [ ] At least 3 external contributors with merged PRs
- [ ] Positive feedback from at least 2 community Go projects

---

### Phase 3: Extension Ecosystem (v1.3.0 → v2.0.0)

**Duration:** 2027 onwards  
**Focus:** Community extensions, advanced features, and ecosystem maturity

#### Deliverables

1. **Community Extension Contract**
   - [ ] Formalize `community-extension.yaml` manifest
   - [ ] Document extension authoring best practices
   - [ ] Registry: `github.com/spcent/plumego-extensions` (index of approved community extensions)

2. **Beta Extensions → GA**
   - [ ] Promote `x/rest`, `x/websocket`, `x/gateway`, `x/observability` to GA as stable roots
   - [ ] Promote `x/tenant`, `x/messaging`, `x/frontend` to GA
   - [ ] All beta packages must have zero API changes across two production cycles

3. **Tooling Enhancements**
   - [ ] `plumego generate spec` — OpenAPI 3.1 generation
   - [ ] `plumego new service` — Scaffold a new Plumego service
   - [ ] `plumego migrate from:chi/gin/echo` — Assisted migration tool
   - [ ] `plumego compat check` — Validate middleware and middleware stacks

4. **Performance & Benchmarks**
   - [ ] Publish updated benchmarks comparing Plumego, Chi, Gin, Echo
   - [ ] Document performance trade-offs and optimization paths
   - [ ] Publish latency and throughput targets for next major version

#### Validation

- [ ] At least 5 community-authored extensions listed in registry
- [ ] Zero API changes in GA packages across release cycle
- [ ] Benchmark results demonstrate competitive performance

---

## Risks and Mitigation

### Risk 1: Positioning Confusion

**Risk:** Plumego might be perceived as "just Chi" or "minimal framework," losing unique positioning.

**Mitigation:**
- Emphasize stdlib compatibility consistently (in README, talks, blog posts).
- Publish side-by-side comparisons: "Plumego vs. Chi," "Plumego vs. Gin."
- Document explicit use cases where Plumego's approach is better.

### Risk 2: Adoption Friction

**Risk:** Developers find Plumego unfamiliar or harder than familiar frameworks.

**Mitigation:**
- Provide step-by-step adoption path (5-min, 30-min, 1-day milestones).
- Create migration guides for developers coming from Chi, Gin, Echo.
- Host community Q&A to answer adoption questions.

### Risk 3: Community Extension Fragmentation

**Risk:** Too many third-party extensions, incompatible with each other or with core.

**Mitigation:**
- Define clear extension authoring contract in `docs/guides/extension-authoring.md`.
- Establish extension review process before listing in registry.
- Use boundary checks to prevent dependency violations.

### Risk 4: API Stability Drift

**Risk:** Stable roots accidentally change, breaking compatibility promise.

**Mitigation:**
- Automate `public-entrypoints-sync` check (current: runs in CI).
- Require API snapshot review before merging changes to stable roots.
- Publish deprecation inventory regularly.

### Risk 5: Documentation Rot

**Risk:** Examples and guides become stale as API evolves.

**Mitigation:**
- Examples are code, compile in CI; failing examples fail the build.
- Link examples to stable root version in docs.
- Review and refresh migration guides quarterly.

---

## Success Criteria

### v1.2.0 (End of Q3 2026)

- [ ] `README.md` refreshed with stdlib-first positioning
- [ ] `STABILITY.md`, `COMPATIBILITY.md`, `docs/EXTENSION_MATURITY.md` published
- [ ] At least 3 working examples in `examples/`
- [ ] All documentation matches current API (no drift)
- [ ] Zero test failures, zero vet warnings

### v1.3.0 (End of Q4 2026)

- [ ] At least 50 GitHub stars
- [ ] At least 3 external contributors with merged PRs
- [ ] Positive feedback from 2+ community Go projects
- [ ] Zero breaking changes to stable roots across v1.1.0–v1.3.0

### v2.0.0 (2027 onwards)

- [ ] At least 3 beta extensions promoted to GA (stable roots)
- [ ] At least 5 community-authored extensions listed in registry
- [ ] Competitive or superior benchmarks vs. Chi, Gin, Echo
- [ ] 10+ production services documented as using Plumego

---

## v1 Stable Surface

The v1 stable surface consists of:

### Stable Roots (GA)

1. `core` — App construction, routing, lifecycle
2. `router` — Route matching, parameters, groups
3. `contract` — Response helpers, error builders
4. `middleware` — Middleware composition
5. `security` — Auth, JWT, password, headers
6. `store` — Storage contracts and in-memory primitives
7. `health` — Health checks
8. `log` — Logging interface and default logger
9. `metrics` — Metrics interface and collectors

### Beta Extensions

1. `x/rest` — CRUD conventions
2. `x/websocket` — WebSocket hub
3. `x/gateway` — Reverse proxy and rewrite
4. `x/observability` — Observability exporters
5. `x/tenant` — Multi-tenancy
6. `x/messaging` — Messaging service
7. `x/frontend` — Static asset serving

### Experimental Extensions

All remaining `x/*` packages; subject to API changes without notice.

---

## Adoption Path

### 5 Minutes

1. Read `README.md` (this document)
2. Run `examples/hello/`
3. Review `docs/start/getting-started.md`

**Goal:** Understand what Plumego is and run a "hello world."

### 30 Minutes

1. Review `reference/standard-service/`
2. Choose one capability from the table in `README.md`
3. Read the corresponding `x/<package>/README.md` primer
4. Implement a small addition to the hello world

**Goal:** Understand the canonical app layout and add one extension.

### 1 Day

1. Read `AGENTS.md` for development workflow
2. Read `STABILITY.md` for compatibility guarantees
3. Review `docs/start/adoption-path.md` for detailed guidance
4. Check `docs/reference/canonical-style-guide.md` for patterns
5. Build a small service from scratch

**Goal:** Understand Plumego's design philosophy and machine-checkable boundaries.

---

## Verification Matrix

| Verification | Command | Baseline | Success |
|---|---|---|---|
| Unit tests | `make gates` | 100% pass | 100% pass |
| Code coverage | `go test -cover ./...` | ≥70% | ≥70% |
| Vet | `go vet ./...` | 0 issues | 0 issues |
| Examples compile | `go build ./examples/...` | all compile | all compile |
| Docs sync | `make website-sync` && git diff | no diff | no diff |
| API stability | `go run ./internal/checks/public-entrypoints-sync` | no drift | no drift |
| Extension maturity | `go run ./internal/checks/extension-maturity` | no drift | no drift |

---

## Decision Log

### June 2026: Documentation & Governance Focus

**Decision:** Prioritize documentation and governance clarity over new features in v1.2.0.

**Rationale:**
- Adoption barriers are primarily educational, not technical.
- Clear stability guarantees reduce risk perception.
- Machine-checkable specs enable community contribution and AI-assisted development.
- Documentation investment has high return for minimal code change.

**Status:** Active. This plan is the implementation vehicle.

---

## Timeline

| Milestone | Target | Status |
|---|---|---|
| Phase 1 documentation complete | Q3 2026 | In progress |
| Phase 1 validation complete | End Q3 2026 | Planned |
| Phase 2 community engagement | Q3–Q4 2026 | Planned |
| Phase 2 adoption metrics | End Q4 2026 | Planned |
| Phase 3 extension ecosystem | 2027 onwards | Planned |
| v2.0 with GA extensions | 2027 H2 | Planned |

---

## Contact and Feedback

For questions about this plan:

1. Open an issue on [GitHub](https://github.com/spcent/plumego).
2. Discuss in the Go community forums or Go Slack.
3. Submit feedback via email to sporengine@gmail.com.

**This plan is a living document.** Updates and adjustments will be published in release notes and reflected in updated versions of this file.
