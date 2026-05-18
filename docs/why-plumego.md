# Why Plumego

Plumego is not a Gin replacement. It is not a faster Echo. It does not compete
on request throughput benchmarks or middleware count.

Release benchmark evidence is recorded separately in
`docs/benchmarks/results-v1.1.0.md`; those numbers are for regression context,
not the adoption claim.

Plumego is a toolkit for teams that need their Go services to remain
understandable and auditable after the first hundred pull requests — and for
codebases where both human engineers and AI coding agents share maintenance
responsibility.

---

## The Problem This Solves

Go web frameworks solve the initial-setup problem well. The hard problem is what
happens after:

- A new engineer joins. Where are the routes? Which middleware runs? Who owns
  which dependency? Most frameworks answer: "trace the framework conventions."
- A pull request touches five packages. Where does each change belong? Most
  frameworks answer: "it depends on how the author learned to use it."
- An AI coding agent is asked to add an endpoint. Which paths are safe to
  modify? Which are frozen? What checks must pass? Most frameworks have no
  machine-readable answer.

Plumego's design addresses all three questions with the same mechanism:
**everything is explicit and locatable in code**.

---

## The Core Principle

> We do not reject dependencies for purity. We isolate dependencies to protect
> the stable kernel.

"Stdlib-only" is not a technical ideology. It is an engineering governance
decision. A stable kernel that carries no third-party imports can be audited,
replaced, and maintained independently of the ecosystem's upgrade cycle.

The stable roots (`core`, `router`, `contract`, `middleware`, `security`,
`store`, `health`, `log`, `metrics`) exist to hold a long-term compatibility
promise. Everything else moves outward into `x/*` capability families, where
change is expected and the compatibility promise is correspondingly lighter.

---

## Stable Root Boundaries

Each stable root has a narrow, explicit scope. These boundaries are enforced by
`internal/checks/dependency-rules`.

| Package | Owns | Does not own |
|---|---|---|
| `core` | App lifecycle, server assembly, graceful shutdown, middleware attachment | DI containers, config loading, ORM, task scheduling |
| `router` | Route matching, path params, group nesting, reverse routing | Controller scanning, annotation routing, response formatting |
| `contract` | Transport-level response/error helpers, request metadata, context accessors, request binding | Business DTOs, domain error types, service injection |
| `middleware` | Transport-level cross-cutting: logging, recovery, timeout, CORS, auth adapters, rate-limiting | Business authorization, tenant policy, domain logic, service lookups |
| `security` | JWT, password utilities, security header policies, input safety, abuse-guard primitives | Full account systems, OAuth platforms, session storage |
| `store` | Storage contracts and idempotency records | ORM, query builders, migration runners, connection pooling |
| `health` | Health and readiness check models, check registration | HTTP handler ownership, external orchestration, service-mesh sidecar policy |
| `log` | Structured logging contracts and default logger | Log aggregation backends, log shipping, cloud-provider logging SDKs |
| `metrics` | Metrics contracts and basic collectors | Prometheus/OpenTelemetry exporters, dashboards, alert rules |

Before adding anything to a stable root, ask three questions:

1. Does this capability belong in the stable kernel, or can it live in `x/*`?
2. If it enters the stable root, can we carry a three-year compatibility promise?
3. Does it pull in a third-party dependency that the stable root does not
   already have?

If any answer is "no" or "uncertain", start in `x/*`.

---

## Explicit Wiring vs. Framework Magic

Plumego requires the caller to write the wiring. This is intentional.

**What Plumego rejects:**

```go
// Global registration — avoid
func init() {
    framework.RegisterController(UserController{})
}

// Auto-scanned routes — avoid
app := framework.Default()
app.AutoRoute("./controllers/...")

// Hidden lifecycle — avoid
framework.Run(":8080")
```

**What Plumego requires:**

```go
// Constructor injection — explicit
logger := plog.NewLogger(cfg.Log)
app := core.New(cfg.Core, core.AppDependencies{
    Logger: logger,
    Store:  store.New(cfg.DB),
})

// Visible middleware order — explicit
app.Use(requestid.Middleware())
app.Use(recovery.Middleware(recovery.Config{Logger: logger}))
app.Use(accesslog.Middleware(accesslog.Config{Logger: logger}))

// Route registration in one file — explicit
app.Get("/ping",  handlers.Ping)
app.Post("/users", handlers.CreateUser)

// Caller-owned shutdown — explicit
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()
if err := app.Shutdown(ctx); err != nil {
    log.Printf("shutdown error: %v", err)
}
```

The longer form pays off when:

- A new engineer can read `internal/app/routes.go` and know every public
  endpoint in under a minute.
- A reviewer can see the exact middleware stack for any route without reading
  framework source.
- An AI agent can modify a route without risk of breaking hidden registration
  side effects.

---

## Extension Families Are Not a Plugin Market

`x/*` packages are **capability families with explicit maturity levels**, not a
plugin ecosystem where all packages carry equal weight. Each family has a status
(`experimental`, `beta`, or `ga`), an owner, and a defined validation path
before promotion.

See `docs/architecture/extension-boundary.md` for the full taxonomy and
promotion criteria.

The practical consequence: when evaluating whether to adopt an `x/*` package,
check its status in `docs/EXTENSION_MATURITY.md` before checking whether it
exists.

---

## Reference Applications Are Part of the Framework

`reference/standard-service` is not a demo. It is the canonical answer to "what
should a Plumego application look like?" Every scaffold, getting-started guide,
and architecture decision points back to it.

Additional reference applications (`reference/with-rest`, `reference/with-gateway`,
`reference/with-websocket`, etc.) demonstrate how to extend the standard shape
for specific capability families, using the same explicit wiring pattern.

See `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md` for how the reference
layer fits into the overall repository structure.

---

## Release Posture Is Evidence, Not Marketing

Plumego ships release evidence alongside every tagged release:

- stable root API snapshots checked against the previous release
- boundary checks (dependency rules, module manifests, reference layout)
- race tests
- extension maturity check output

This means the compatibility claim is verifiable, not assumed. See
`docs/release/PRE_V1_RELEASE_CHECKLIST.md` for the full evidence model.

---

## AI Agent Maintainability

Plumego's repository is designed for codebases where AI coding agents participate
in maintenance alongside human engineers. Machine-readable specs, per-module
manifests, standardized check commands, and scoped task cards give agents a
clear operating model.

See `docs/agent-first.md` for the full agent-first design rationale and
mechanism reference.

---

## Fit Summary

Plumego is the right choice when:

- Route registration, middleware order, and dependency wiring must be visible in
  code review without reading framework source.
- Multiple services need to start from one canonical shape that is easy to hand
  off between teams.
- Adoption decisions must be based on explicit maturity evidence, not package
  availability.
- The codebase will be maintained by a mix of engineers and AI coding agents
  over a multi-year horizon.

See `docs/when-not-to-use-plumego.md` for cases where Plumego is the wrong
choice.
