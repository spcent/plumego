---
title: "Plumego v1.0 — A Stable, Explicit Foundation for Long-Lived Go Services"
slug: "plumego-v1-release-announcement"
date: 2026-01-10T10:00:00+01:00
lastmod: 2026-01-10T10:00:00+01:00
description: "Plumego v1.0 is officially released. This announcement explains what v1.0 means, what Plumego deliberately is and is not, and why this release is about stability and trust—not features."
keywords:
  - plumego
  - v1.0
  - golang framework
  - release announcement
  - birdor
tags:
  - Go
  - Framework
  - Release
  - Architecture
categories:
  - Engineering
toc: false
---

# A Stable, Explicit Foundation for Long-Lived Go Services

## Plumego v1.0 is released

Today we are releasing **Plumego v1.0**.

This release does not mark the moment when Plumego became *feature-complete*.
It marks the moment when Plumego became **semantically stable**.

From this version forward, Plumego makes a clear promise:

> The core behavior you rely on today will still behave the same years from now.

That promise is the real meaning of v1.0.

---

## What Plumego v1.0 means

Plumego v1.0 means:

- The public API surface is stable
- Core runtime semantics are frozen
- Architectural boundaries are explicit and documented
- Breaking changes require a new major version
- The framework is safe to adopt for long-lived systems

It does **not** mean:

- Plumego now does everything
- Plumego is trying to replace application platforms
- Plumego has chosen convenience over clarity

v1.0 is about **trust**, not scope.

---

## Why Plumego exists

Plumego was created to solve a specific problem we repeatedly encountered in real systems:

> Go web services often start simple,  
> but their frameworks quietly accumulate hidden behavior,  
> until the system becomes hard to reason about and harder to change.

Plumego takes a different approach.

It assumes:
- Systems will grow
- Teams will change
- Requirements will shift
- Code will be read far more often than it is written

And it optimizes for that reality.

---

## Design principles that define v1.0

Plumego v1.0 is shaped by a small set of non-negotiable principles.

### Explicit over implicit

Plumego avoids hidden lifecycle hooks, implicit wiring, and magic defaults.

If something happens:
- it is visible in code
- it is documented
- it is testable

### Standard library first

Plumego builds on `net/http`, `context`, and other standard packages.

This keeps:
- mental overhead low
- debugging familiar
- long-term maintenance predictable

### Stable boundaries, not maximal features

Plumego defines **where responsibilities live**:
- routing
- middleware
- authentication boundaries
- background lifecycle
- error handling

It deliberately does not define everything else.

---

## What is included in v1.0

Plumego v1.0 provides a **complete, stable framework core**:

- Deterministic HTTP runtime
- Explicit router and middleware system
- Clear context usage contract
- Recommended error and response patterns
- Defined authentication and session boundaries
- Background task lifecycle integration
- Observability and debugging baselines
- Production-ready examples and documentation

Every included capability is:
- documented
- tested
- accompanied by real examples

---

## What Plumego intentionally does not include

To be clear, Plumego v1.0 does **not** include:

- An ORM
- A dependency injection container
- A configuration framework
- An admin panel
- A plugin marketplace

These are not omissions.
They are deliberate boundaries.

Plumego aims to be the **foundation you build on**, not the platform that decides for you.

---

## Who Plumego v1.0 is for

Plumego v1.0 is
