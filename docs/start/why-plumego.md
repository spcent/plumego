# Why Plumego

Plumego is not a Gin replacement. It is not a faster Echo. It does not compete
on request throughput benchmarks or middleware count.

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
`internal/checks/dependency-rules`. For detailed ownership tables, see
`docs/concepts/core-boundary.md`.

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

// Caller-owned lifecycle — explicit
if err := app.Prepare(); err != nil {
    log.Fatal(err)
}
srv, err := app.Server()
if err != nil {
    log.Fatal(err)
}
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()
go func() {
    <-ctx.Done()
    _ = app.Shutdown(context.Background())
}()
if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
    log.Printf("server stopped: %v", err)
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

See `docs/concepts/extension-boundary.md` for the full taxonomy and
promotion criteria.

The practical consequence: when evaluating whether to adopt an `x/*` package,
check its status in `docs/concepts/extension-maturity.md` before checking whether it
exists.

---

## Reference Applications Are Part of the Framework

`reference/standard-service` is not a demo. It is the canonical answer to "what
should a Plumego application look like?" Every scaffold, getting-started guide,
and architecture decision points back to it.

Additional reference applications (`reference/with-rest`, `reference/with-gateway`,
`reference/with-websocket`, etc.) demonstrate how to extend the standard shape
for specific capability families, using the same explicit wiring pattern:
`run()` owns process signals and server lifecycle, `app.Prepare()` freezes
routes, and business examples live under `internal/domain/<name>`.

See `AGENTS.md §3` for how the reference layer fits into the overall
repository structure.

---

## Release Posture Is Evidence, Not Marketing

Plumego ships release evidence alongside every tagged release:

- stable root API snapshots checked against the previous release
- boundary checks (dependency rules, module manifests, reference layout)
- race tests
- extension maturity check output

This means the compatibility claim is verifiable, not assumed. See
`docs/release/v1.0.0.md` and `docs/release/v1.1.0.md` for the release evidence.

---

## AI Agent Maintainability

Plumego's repository is designed for codebases where AI coding agents participate
in maintenance alongside human engineers. Machine-readable specs, per-module
manifests, standardized check commands, and scoped task cards give agents a
clear operating model.

See `AGENTS.md` for the full agent-first design rationale and mechanism
reference.

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

See `docs/start/when-not-to-use-plumego.md` for cases where Plumego is the wrong
choice.
