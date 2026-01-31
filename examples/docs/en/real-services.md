# Plumego for Real Services

## A production-grade Go web framework for systems that must survive real life

Plumego is a **standard-library-first Go web framework** designed for services that are expected to run, evolve, and remain understandable for years.

It focuses on the parts that real systems always need —  
**security, lifecycle, integrations, and observability** —  
without locking you into a platform or hiding behavior behind magic.

---

## The problem with “getting started fast”

Most Go web frameworks optimize for the first week:

- fast routing
- quick middleware setup
- easy demos

But real services don’t fail in the first week.

They fail months later, when teams have added:
- ad-hoc auth logic
- inconsistent middleware
- fragile shutdown behavior
- copy-pasted webhook handlers
- observability bolted on too late

By then, the framework is no longer helping — it’s just in the way.

**Plumego is built for what happens after that.**

---

## What Plumego provides (for real services)

### A stable service runtime
Plumego ships with runtime guarantees that most teams eventually re-implement:

- graceful startup and shutdown  
- concurrency limits and backpressure  
- request timeouts and body size limits  
- panic recovery  
- readiness and health endpoints  

These are not examples or plugins.  
They are **part of the runtime**.

---

### Security as a first-class boundary
Plumego treats security as architecture, not middleware glue.

- explicit authentication boundaries  
- authorization contracts (roles, scopes, policies)  
- access + refresh token patterns  
- session tracking and revocation  
- clear 401 vs 403 semantics  

Security is designed to be **auditable, testable, and evolvable**.

---

### Integrations built into the story
Modern services are integration-heavy.

Plumego includes:
- inbound webhook handling with verification and replay protection  
- outbound webhook delivery and retry  
- in-process Pub/Sub for event fan-out  
- realtime utilities for notifications and dashboards  

These are everyday business needs — not edge cases.

---

### Observability without lock-in
Plumego provides observability through **interfaces and adapters**, not mandates.

- request ID propagation  
- metrics adapters (Prometheus)  
- tracing adapters (OpenTelemetry)  
- structured logging compatibility  

Start simple. Grow safely. Change vendors without rewriting your runtime.

---

### Explicit composition, no hidden magic
Plumego is built directly on Go’s standard library.

- `http.Handler` remains the core abstraction  
- middleware order is code order  
- lifecycle is explicit  
- no reflection-driven DI  
- no hidden global state  

If you can read Go code, you can understand the system.

---

## Where Plumego fits best

### User Centers & Auth Services
- login / refresh / revoke  
- multi-device sessions  
- RBAC and policy enforcement  
- long-term security guarantees  

### SaaS Backends
- multi-tenant isolation  
- background jobs and lifecycle hooks  
- webhooks and third-party integrations  
- safe rollouts and shutdowns  

### Platform & Integration Services
- inbound/outbound webhooks  
- event fan-out  
- burst traffic handling  
- observability-first design  

### Internal Platforms / Middle Offices
- consistent runtime across services  
- predictable behavior  
- low onboarding cost  
- long-term maintainability  

---

## How Plumego compares

Plumego is not trying to replace everything.

- More complete than a bare router
- Less opinionated than a full application platform
- More predictable than convenience-first frameworks

It sits in the space where **real services actually live**.

---

## Who Plumego is for

Plumego is a strong choice if you:

- build services expected to last years  
- care about operational safety  
- value explicit architecture and code review  
- want control without reinventing the runtime  

It may not be the best choice if you want:
- maximum scaffolding
- a batteries-included platform
- minimal upfront decisions

Both are valid paths. Plumego is explicit about which one it takes.

---

## Stability you can rely on

Plumego follows semantic versioning.

Starting with v1.0:
- core behavior is stable  
- contracts are documented  
- breaking changes require a major version  
- examples are treated as part of the API  

Upgrades should never feel surprising.

---

## Final thought

Plumego does not try to impress you with features.

It tries to earn your trust by remaining understandable when systems grow, teams change, and incidents happen.

If you are building **real services** —  
services that must be secure, observable, and maintainable —

**Plumego is built for that reality.**

---

### Get started
- Read the documentation  
- Explore real-world examples  
- Understand the contracts  

Plumego is calm by design.  
So your services don’t have to be.
