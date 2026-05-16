# When Not to Use Plumego

Plumego makes deliberate trade-offs. Some of those trade-offs are wrong for
certain teams and projects. This document is direct about those cases.

Read this before investing time in the technical details. If any section below
describes your situation, a different tool is the better fit.

---

## You want the framework to manage service structure automatically

Plumego requires the caller to write explicit bootstrap, route registration, and
middleware wiring. If your team's preference is for the framework to handle
startup, inject dependencies, and organize routes through conventions or
annotations, Plumego will feel unnecessarily deliberate.

Frameworks like Gin, Echo, and similar convention-first toolkits are better fits
for teams that want structure to be managed for them.

---

## Every service has its own intentionally custom structure

Plumego's reference application (`reference/standard-service`) defines one
canonical layout. The toolkit is strongest when multiple services start from that
shared baseline.

If each service in your organization is intentionally built with its own startup
pattern, its own route organization, and its own dependency wiring approach,
Plumego's shared canonical path adds friction without a corresponding benefit.

---

## You want to use packages without following a structured adoption model

Plumego distinguishes between stable roots, the canonical reference app,
scenario references, and experimental extension families. That distinction
matters: it determines what can be adopted for production use and what requires
caution.

If your team does not need to distinguish between stable and experimental
modules — if package availability is sufficient signal for adoption — then
Plumego's maturity discipline will feel like unnecessary overhead.

---

## You need a large ecosystem of ready-made integrations

Plumego's stable kernel is intentionally narrow, and its `x/*` extension
families cover the most common capability patterns. But Plumego does not have
the broad middleware and integration ecosystem of Gin, Echo, or Fiber.

If the speed of reaching for a ready-made community package matters more than
kernel stability and wiring visibility, a framework with a larger ecosystem is a
better starting point.

---

## Your primary concern is maximum raw throughput

Plumego is built on `net/http` and does not sacrifice correctness or
explicitness for throughput. If your service is a high-traffic edge proxy or
gateway where raw request-per-second performance is the primary constraint, a
lower-level toolkit or a framework built specifically for throughput
optimization may be a better fit.

This does not apply to most internal APIs and platform services, where the
bottleneck is almost never the HTTP framework.

---

## You are building a short-lived prototype with no maintenance horizon

Plumego's value concentrates over time: explicit wiring pays off at the tenth PR
and the hundredth onboarding. For a throwaway proof-of-concept where the code
will be discarded within weeks, the upfront explicitness cost is not recovered.

---

## Slow-down Signals

Before adopting Plumego, check whether any of these apply:

| Signal | Implication |
|---|---|
| The team wants routing and injection handled invisibly | Plumego will feel more deliberate than productive |
| Services are each intentionally unique | A shared canonical path reduces value |
| Package existence implies production readiness to the team | Plumego's maturity separation adds friction |
| The project needs 30+ community middleware packages immediately | Plumego's ecosystem is narrower |
| It is a one-person script or short-lived spike | Overhead exceeds benefit |

---

## Where to Go Instead

| Need | Better starting point |
|---|---|
| Convention-first, large ecosystem | Gin, Echo |
| Minimal stdlib routing only | Chi |
| Express-style familiarity | Fiber |
| Maximum raw performance | fasthttp-based toolkits |
| Rapid prototyping with batteries included | Any full-featured framework |

These are not inferior choices. They solve different problems. Plumego solves
the long-term explicit-wiring, agent-maintainability problem. If that is not
your problem, use the tool that fits.
