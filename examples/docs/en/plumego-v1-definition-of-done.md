---
title: "Plumego v1.0 — Definition of Done"
slug: "plumego-v1-definition-of-done"
date: 2026-01-01T10:00:00+01:00
lastmod: 2026-01-01T10:00:00+01:00
description: "A precise, engineering-driven Definition of Done for Plumego v1.0. This document defines what must be complete, stable, documented, and verifiable before Plumego can be considered a production-ready Go web framework."
keywords:
  - plumego
  - v1.0
  - definition of done
  - golang framework
  - architecture
  - birdor
tags:
  - Go
  - Framework
  - Roadmap
  - Architecture
categories:
  - Docs
toc: true
---

# Definition of Done

## Purpose of this document

This document defines the **non-negotiable completion criteria** for **Plumego v1.0**.

It is intentionally strict.

Plumego v1.0 is not defined by:
- popularity
- number of features
- marketing readiness

Plumego v1.0 is defined by:
> **semantic stability, architectural clarity, and long-term maintainability guarantees.**

This Definition of Done (DoD) is used to decide:
- whether v1.0 can be released
- whether APIs are considered stable
- whether Plumego is safe to adopt for long-lived systems

---

## v1.0 Scope Statement

Plumego v1.0 is a **production-grade Go web framework core**, not a full application platform.

It guarantees:
- a stable HTTP runtime
- explicit composition
- predictable behavior
- safe extension points

It does **not** guarantee:
- batteries-included SaaS features
- ORM, migrations, admin UI, or cloud integrations
- automatic architecture decisions

---

## Global v1.0 Invariants (Must Hold)

Before v1.0 is released, all of the following must be true:

- All public APIs are explicitly documented
- Core runtime semantics are frozen
- All behaviorally significant components have tests
- Breaking changes require a major version
- Every core feature has a reference example
- The framework can be understood without reading source code

If any invariant is violated, v1.0 must not be released.

---

## 1. Core Runtime (HTTP & Lifecycle)

### Definition of Done

- [ ] `core.App` lifecycle is fully specified and documented
- [ ] Startup order is deterministic and documented
- [ ] Graceful shutdown semantics are defined:
  - context cancellation
  - timeout behavior
  - order of component shutdown
- [ ] Panic recovery behavior is stable and tested
- [ ] `core.App` implements `http.Handler` without hidden side effects

### Required Artifacts

- `docs/architecture/core-lifecycle.md`
- Unit tests for startup/shutdown edge cases
- Example demonstrating graceful shutdown

---

## 2. Router (Matching & Parameters)

### Definition of Done

- [ ] Route matching precedence is explicitly documented
- [ ] Parameter extraction rules are fixed and tested
- [ ] URL decoding behavior is specified
- [ ] 404 vs 405 behavior is stable and documented
- [ ] No undocumented routing side effects exist

### Required Artifacts

- `docs/architecture/router-contract.md`
- Golden tests for route matching
- Benchmark for route lookup performance

---

## 3. Middleware System

### Definition of Done

- [ ] Middleware execution order is deterministic
- [ ] Short-circuit semantics are documented
- [ ] Error propagation rules are explicit
- [ ] Middleware chaining does not allocate excessively
- [ ] Middleware can be tested in isolation

### Required Artifacts

- `docs/architecture/middleware-contract.md`
- Golden tests for middleware order
- Example middleware stack

---

## 4. Context Usage Contract

### Definition of Done

- [ ] Allowed vs forbidden context usage is documented
- [ ] Context key namespace rules are defined
- [ ] Cancellation and timeout behavior is tested
- [ ] No framework code abuses context as a data store

### Required Artifacts

- `docs/architecture/context-contract.md`
- Tests validating context cancellation propagation

---

## 5. Error & Response Model (Recommended Path)

### Definition of Done

- [ ] Official recommended error classification exists
- [ ] HTTP status mapping rules are documented
- [ ] Example response envelope is defined
- [ ] Error handling does not rely on panic for control flow

### Required Artifacts

- `docs/best-practices/error-handling.md`
- Example handlers demonstrating error flow
- Tests for common error paths

---

## 6. Auth & Security (Framework Boundary + Example)

### Definition of Done

- [ ] Official Auth boundary is defined (contract-first)
- [ ] `Principal` structure is frozen
- [ ] `Authenticator`, `Authorizer`, `SessionStore` contracts are frozen
- [ ] `RequireAuth` middleware semantics are stable
- [ ] Refresh token rotation and revocation are demonstrated
- [ ] Security baseline is documented

### Required Artifacts

- `docs/security/auth-contract.md`
- `docs/security/_index.md`
- `examples/user-center` (register/login/refresh/logout)
- Tests covering auth edge cases

---

## 7. Background Tasks & Lifecycle Integration

### Definition of Done

- [ ] Background runner interface is defined
- [ ] Registration and shutdown order is deterministic
- [ ] Failure handling strategy is documented
- [ ] Background tasks respect context cancellation

### Required Artifacts

- `docs/architecture/background-tasks.md`
- Example background worker
- Shutdown tests

---

## 8. Observability Baseline

### Definition of Done

- [ ] Request ID generation is standardized
- [ ] Context propagation rules are documented
- [ ] Logging field conventions are recommended
- [ ] Debug/diagnostic output is available

### Required Artifacts

- `docs/observability/_index.md`
- Request ID middleware example
- Diagnostic dump or `doctor` output example

---

## 9. Configuration & Startup Model

### Definition of Done

- [ ] Recommended configuration pattern is documented
- [ ] Environment override strategy is clear
- [ ] Startup phases are explicit
- [ ] Misconfiguration failure modes are documented

### Required Artifacts

- `docs/best-practices/configuration.md`
- Example config struct and loader

---

## 10. Examples & Learning Path

### Definition of Done

- [ ] At least one complete, real-world example exists
- [ ] Example covers auth, routing, middleware, lifecycle
- [ ] Examples are runnable and tested
- [ ] Docs reference examples consistently

### Required Artifacts

- `examples/user-center`
- Example-as-test coverage
- Getting Started walkthrough

---

## 11. Documentation Completeness

### Definition of Done

- [ ] Docs explain *why*, not only *how*
- [ ] Explicit non-goals are documented
- [ ] FAQ covers common misconceptions
- [ ] Roadmap beyond v1.0 exists

### Required Artifacts

- `docs/_index.md`
- `docs/faq/*`
- `docs/roadmap/*`

---

## 12. Stability & Release Guarantees

### Definition of Done

- [ ] Public API surface is enumerated
- [ ] Deprecation policy is documented
- [ ] Migration guide template exists
- [ ] Semantic versioning rules are enforced

### Required Artifacts

- `MIGRATIONS.md`
- `CHANGELOG.md`
- API snapshot or equivalent verification

---

## Final Release Gate

Plumego v1.0 **may only be released** when:

- All sections above are complete
- All required artifacts exist
- All tests pass
- All contracts are documented
- Maintainers agree no undefined behavior remains

---

## Final Statement

> **Plumego v1.0 is not about adding more features.  
It is about removing uncertainty.**

When this Definition of Done is satisfied, Plumego can confidently claim:

- stability
- predictability
- long-term engineering value

Anything less is still a pre-1.0 experiment.
