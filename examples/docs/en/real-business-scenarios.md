## Real Business Scenarios

Plumego is designed around how real services actually evolve.
Below are four common production scenarios where Plumego fits naturally — without forcing a platform or hiding complexity.

---

### 1. User Center / Authentication Service

**Typical requirements**
- User registration and login
- Access & refresh tokens
- Session revocation and forced logout
- Multi-device login
- Role / permission checks
- Auditability and security hardening

**How Plumego helps**
- Clear auth boundaries (`Authenticator`, `Authorizer`, `SessionStore`)
- Explicit middleware semantics (401 vs 403, injection rules)
- Refresh token rotation and revocation patterns
- Security defaults: timeouts, body limits, headers
- Long-lived session management without global magic

**Why Plumego fits**
User centers tend to outlive most other services.
Plumego makes security and lifecycle explicit, so auth logic remains auditable and evolvable over time.

**Typical Plumego stack**
- HTTP runtime + security middleware
- Auth middleware + session store
- Role-based authorization
- Health and readiness endpoints

---

### 2. SaaS Backend (Multi-Tenant)

**Typical requirements**
- Tenant isolation
- Role-based access control
- Webhooks to/from third parties
- Background jobs (billing, sync, cleanup)
- Observability across tenants
- Safe rollout and graceful shutdown

**How Plumego helps**
- Tenant-aware `Principal` and context propagation
- Contract-first auth and authorization
- Inbound & outbound webhook support
- Pub/Sub for internal domain events
- Explicit lifecycle hooks for background tasks
- Metrics and tracing adapters without vendor lock-in

**Why Plumego fits**
SaaS backends grow in surface area over time.
Plumego provides a stable runtime foundation while letting teams evolve storage, billing, and domain logic independently.

**Typical Plumego stack**
- Core runtime + lifecycle hooks
- Auth + tenant-aware authorization
- Webhooks + Pub/Sub
- Observability adapters

---

### 3. Platform / Integration Service

**Typical requirements**
- Webhook receivers (GitHub, Stripe, internal systems)
- Signature verification and replay protection
- Event fan-out
- Debugging and replay tooling
- High reliability under burst traffic

**How Plumego helps**
- Built-in inbound webhook handlers with validation
- Outbound webhook delivery and replay
- In-process Pub/Sub for fan-out
- Concurrency limits and backpressure
- Explicit error handling and observability

**Why Plumego fits**
Integration services are failure-prone and hard to debug.
Plumego treats integrations as first-class runtime concerns instead of ad-hoc handlers.

**Typical Plumego stack**
- Webhook components
- Pub/Sub
- Metrics + tracing
- Strict runtime guards

---

### 4. Internal Platform / Middle Office Services

**Typical requirements**
- Multiple internal APIs
- Consistent security and auth
- Operational safety
- Clear ownership boundaries
- Long-term maintainability
- Easy onboarding for new team members

**How Plumego helps**
- Explicit composition makes service behavior easy to audit
- Stable runtime semantics reduce tribal knowledge
- Standardized health, logging, and observability
- No heavy framework lock-in
- Clear contracts between layers and teams

**Why Plumego fits**
Internal platforms suffer most from unclear architecture and hidden behavior.
Plumego optimizes for clarity, reviewability, and predictable evolution.

**Typical Plumego stack**
- Core runtime
- Shared auth and middleware
- Background runners
- Observability baseline

---

## A common pattern across all scenarios

Across user centers, SaaS backends, platforms, and internal services, Plumego consistently provides:

- A **stable, production-grade runtime**
- Explicit security and lifecycle boundaries
- Integration primitives that real systems need
- Observability without forcing early decisions
- A foundation that does not need to be rewritten as the system grows

Plumego does not try to define your business.
It ensures that your business runs on a **safe, explicit, and durable foundation**.

---

## When Plumego may not be the best choice

Plumego may not be ideal if:
- You want a batteries-included application platform
- You expect scaffolding to define your architecture
- You optimize primarily for short-lived demos

Plumego is built for systems that will still matter years from now.
