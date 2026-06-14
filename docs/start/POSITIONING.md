# Why Plumego?

Plumego is the standard-library-first HTTP toolkit for Go developers building explicit, maintainable, agent-friendly services.

## The Design Philosophy

Plumego combines the familiar shapes of `net/http` with just enough structure to build production services **without a framework**. Handlers are plain `func(http.ResponseWriter, *http.Request)`. Middleware wraps `http.Handler`. Wiring is visible in your `main` package. The stable core has zero external dependencies.

This is intentional: we optimize for **understanding**, **testing**, and **long-term maintenance** over convenience.

## Choose Plumego if you...

- **Want to understand every line of your HTTP server's wiring.** No hidden binding, no context service-locators, no reflection-based DI.
- **Prefer stdlib shapes and patterns.** You already know `net/http`, `httptest`, request context accessors, and middleware composition.
- **Expect your service to live for years with predictable maintenance.** Code should read the same way next year as it does today.
- **Use code agents (Claude, Codex) to assist development.** Plumego's control plane (`specs/`, `reference/`, `module.yaml`) makes scope and validation discoverable.
- **Value small, testable, refactorable code over convenience.** You're willing to write more wiring code in exchange for fewer surprises.

## What Plumego is NOT

- **Not a "Gin/Echo replacement."** We're not chasing their DevX or speed. We're complementary to stdlib, not competitive with frameworks.
- **Not the fastest.** (Fiber optimizes for raw throughput; we optimize for clarity.)
- **Not "batteries-included."** We're stdlib-first by design; rich features live outside the stable learning path.
- **Not for teams who want zero wiring code.** If you prefer `plumego.Handlers{ "POST /users": createUser }` over explicit routes, you need a different framework.

## How Plumego compares

| Toolkit | Position | Tradeoff | Best for |
|---------|----------|----------|----------|
| `http.ServeMux` (stdlib) | Minimal routing | Zero framework learning; high wiring boilerplate | Trivial services, learning |
| **Plumego** | **Thin layer on stdlib** | **Small API; explicit wiring; full stdlib compatibility** | **Production services with stable maintenance** |
| Chi | Lightweight router | Pleasant router API; diverges from stdlib middleware patterns | Developers who like function-builder style |
| Gin | Fast + convenient | Powerful DevX; hides binding, response wrapping, panic handling | High-velocity prototyping |
| Echo | Feature-rich | Broad conventions; less transparent; steeper learning curve | Fully-featured applications with batteries |
| Fiber | High performance | Non-stdlib (fasthttp); significant framework coupling | Maximum throughput, benchmark-oriented |

## Start here

- **5 minutes:** Run `examples/hello` to see it work
- **30 minutes:** Copy `reference/standard-service`, add one endpoint
- **1 day:** Read `AGENTS.md` and understand the control plane

Then pick the section that matches your need:

| I want to build... | Start here |
|---|---|
| A plain JSON API | `reference/standard-service` |
| REST with CRUD conventions | `reference/with-rest` → `x/rest` (beta) |
| Multi-tenant service | `reference/with-tenant` → `x/tenant` (beta) |
| Real-time WebSocket | `reference/with-websocket` → `x/websocket` (beta) |
| API gateway or proxy | `reference/with-gateway` → `x/gateway` (beta) |
| Messaging/webhooks | `reference/with-messaging` → `x/messaging` (beta) |
| Observability stack | `reference/with-observability` → `x/observability` (beta) |

See `docs/start/adoption-path.md` for the full 2-hour onboarding flow.

## Key principles

1. **Standard library first** — All stable roots use only stdlib. Bring your own ORM, cache, or message queue.
2. **Explicit wiring** — Routes, middleware, dependencies, and lifecycle are visible at construction sites. No hidden registration, no import-order behavior.
3. **Small stable surface** — Nine packages, stable v1 guarantee. Learn the whole thing or just use what you need.
4. **Agent-friendly** — Specs, tasks, reference examples, and per-module documentation make scope and validation discoverable for code agents.
5. **Optional capabilities** — Product features and protocol adapters live in `x/*`, not baked into the core learning path.

## What you get in v1

**9 stable roots** (full v1 compatibility):
- `core` — App construction, lifecycle, route registration
- `router` — Route matching, params, groups, reverse URLs
- `contract` — Response writers, structured errors, transport binding
- `middleware` — Transport-layer middleware composition + standard packages
- `security` — Auth, JWT, password hashing, abuse guards
- `store` — Storage contracts + in-memory primitives
- `health` — Health/readiness models
- `log` — Logging interfaces + default logger
- `metrics` — Metrics contracts + collectors

**7 beta extensions** (stable across release refs, suitable for production):
- `x/rest`, `x/websocket`, `x/gateway`, `x/observability`, `x/tenant`, `x/frontend`, `x/messaging`

**8 experimental extensions** (APIs may change):
- `x/ai`, `x/data`, `x/fileapi`, `x/openapi`, `x/resilience`, `x/rpc`, `x/validate`, and subordinate packages

See `STABILITY.md` for the full compatibility guarantee and `COMPATIBILITY.md` for upgrade paths.

## Maintenance model

Plumego is **agent-first**: maintained using code agents as regular contributors, not as a separate workflow. This means:

- Boundaries are machine-checkable (`specs/dependency-rules.yaml`)
- Tasks are explicit and verifiable (`tasks/milestones/active/`, `tasks/cards/active/`)
- Changes produce evidence, not hidden convention (`internal/checks/`)
- Reference examples are canonical, not aspirational (`reference/standard-service`)

This makes the project **safe for code agents** and **readable for humans** at the same time.

## Next steps

1. Read `docs/start/getting-started.md` for the smallest runnable example
2. Read `docs/start/adoption-path.md` for the 5-min → 30-min → 1-day learning flow
3. Copy `reference/standard-service` as your baseline
4. Join the community to ask questions or propose extensions

---

**Still deciding?** Check `docs/guides/migration/` for guides coming from Gin, Echo, Chi, or plain stdlib.
