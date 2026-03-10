# CORE API Contraction Proposal

> Historical document (superseded).  
> This proposal is archived and is **not** the canonical v1 guidance as of March 10, 2026.  
> Use `docs/other/V1_CORE_API_FREEZE.md` and `docs/CANONICAL_STYLE_GUIDE.md` as the source of truth.

Status: Draft v0.1  
Scope: `core`, `router`, `contract`, `middleware`, canonical examples  
Audience: maintainers, contributors, code agents, reviewers  
Related:
- `docs/AI_FRIENDLY_CORE_GUIDE.md`
- `docs/DEPRECATION_PLAN_v0_TO_v1.md`
- `docs/CANONICAL_STYLE_GUIDE.md`

---

## 1. Purpose

This document turns the principles in the related documents into a concrete API-contraction plan for Plumego.

The goal is not to remove useful capabilities. The goal is to make the **core runtime smaller, more predictable, and easier for AI agents and human contributors to modify without spreading style drift**.

Plumego already has strong foundations:
- standard-library-first runtime
- `http.Handler` compatibility
- middleware built around standard `net/http` semantics
- a clear `core` package as the primary entrypoint

The main issue is not lack of capability. The issue is that the **public core surface is wider than the canonical mental model**. Multiple registration styles, mixed handler forms, and non-routing responsibilities inside `router` increase ambiguity for both humans and agents.

This proposal defines how to contract the public API surface from the current state toward a v1-ready core.

---

## 2. Design Goal

After this contraction, a new contributor or code agent should be able to answer the following questions quickly and correctly:

1. What is the default way to bootstrap an app?
2. What is the default handler signature?
3. What is the default way to decode input?
4. What is the default way to write responses and errors?
5. Which packages are core, and which are optional extensions?
6. Which APIs are compatibility-only and must not be used in new code?

If any of these require reading several packages or examples, the core is still too wide.

---

## 3. Current Problems This Proposal Addresses

### 3.1 Parallel route registration styles

Today, `core.App` exposes parallel route registration styles such as:
- `Get/Post/...`
- `GetCtx/PostCtx/...`
- `GetHandler/PostHandler/...`
- `Handle/HandleFunc`

This creates avoidable ambiguity. A code agent can generate valid code in multiple styles, which leads to repository drift.

### 3.2 More than one “first-class” handler shape

Today, a reader can infer at least three important handler shapes:
- standard `http.HandlerFunc`
- context-aware `CtxHandlerFunc`
- router-specific handler shapes and adapters

This is too many for a small core.

### 3.3 `router` package is carrying non-router responsibilities

The `router` package currently exposes concepts that are not part of a pure routing layer, including resource/controller/repository/validator/JSON-writer concerns.

This breaks package meaning. For AI agents, this is especially harmful because package names are part of how the agent infers architecture.

### 3.4 Optional capabilities feel too close to core

WebSocket, webhook, ops/admin, multi-tenancy, scheduler, and other capability packs are useful, but they are too visible in the same mental space as the core runtime.

### 3.5 No single contracted error/response path

Structured responses exist, but the framework still needs a stricter “one official path” for JSON success and JSON error writing.

---

## 4. Core Contraction Principles

This proposal applies the following rules.

### 4.1 One canonical style, many adapters

Plumego may keep adapters for compatibility, but **only one path may be documented as canonical**.

### 4.2 Core must optimize for predictability, not convenience breadth

If a helper creates a second valid style for common application code, it should be moved to compatibility or extension space.

### 4.3 Package name must match package role

A package named `router` must be about routing. A package named `contract` must be about request/response contracts. A package named `middleware` must be about middleware.

### 4.4 Hidden request-time state must be minimized

Middleware may enrich request context for narrow infrastructure cases, but application DTO flow must stay explicit by default.

### 4.5 New public APIs must compete against contraction goals

A new API is acceptable only if it reduces ambiguity, not if it simply adds another convenient path.

---

## 5. Target Canonical Core Shape

This section defines the target shape after contraction.

### 5.1 Canonical app bootstrap

The canonical bootstrap remains:

> Legacy snippet (historical): uses pre-freeze option-style middleware wiring; not canonical v1.

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithRecovery(),
    core.WithLogging(),
)
```

The canonical serve path is either:

```go
if err := app.Boot(); err != nil {
    log.Fatal(err)
}
```

or explicit standard-library serving:

```go
log.Fatal(http.ListenAndServe(":8080", app))
```

The docs should present one of these first and keep the other as a clearly labeled alternative.

### 5.2 Canonical route registration

Canonical route registration must be:

```go
app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
    // ...
})
```

or, if Plumego decides to make context-style canonical instead:

```go
app.GetCtx("/users/:id", func(ctx *contract.Ctx) {
    // ...
})
```

**Only one of these may remain canonical for v1.**

This proposal recommends the first option:
- standard `http.HandlerFunc` is canonical
- context-style is supported through adapters/utilities
- docs emphasize standard-library-first routing

### 5.3 Canonical handler contract

Recommended canonical handler contract for core:

```go
type HandlerFunc = http.HandlerFunc
```

Context-aware handlers remain supported but should be classified as one of:
- adapter path
- convenience path
- compatibility path

They should not remain a parallel first-class style in docs and examples.

### 5.4 Canonical input decoding

Canonical input decoding must be explicit.

Preferred style:

```go
var req CreateUserRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    contract.WriteError(w, r, contract.ErrBadRequest("invalid_json"))
    return
}
```

Allowed thin helper:

```go
if err := contract.DecodeJSON(r, &req); err != nil {
    contract.WriteError(w, r, contract.ErrBadRequest("invalid_json"))
    return
}
```

Not canonical:
- hidden middleware-based DTO injection
- multi-source autobind from path/query/header/body in one implicit step
- “magic bind and validate” in general-purpose handlers

### 5.5 Canonical response path

Core must standardize one success path and one error path:

```go
contract.WriteResponse(w, r, http.StatusOK, payload)
contract.WriteError(w, r, err)
```

Context helpers may delegate to the same internals, but there must be only one contract layer behind them.

### 5.6 Canonical package boundary

Target core packages:
- `core` — app bootstrap, lifecycle, composition
- `router` — route matching, params, groups, freeze, mount behavior
- `middleware` — middleware primitives and standard middleware
- `contract` — request/response contracts, decode helpers, response helpers, error model

Target non-core packages:
- `components/...`
- `websocket`
- `webhook`
- `tenant`
- `scheduler`
- `ops`
- `frontend`
- `ai`
- any persistence/resource/controller package

---

## 6. Public API Contraction Decisions

This section proposes concrete decisions.

## 6.1 Route registration API

### Decision

Keep one primary route registration family. All others become compatibility or adapter APIs.

### Recommended v1 primary family

Keep:
- `Get`
- `Post`
- `Put`
- `Patch`
- `Delete`
- `Head`
- `Options`
- `Handle`

Primary expected handler shape:
- `http.Handler`
- `http.HandlerFunc`

### De-emphasize

Move to compatibility-only or adapter-only status:
- `GetCtx/PostCtx/...`
- `GetHandler/PostHandler/...`
- other parallel route registration helpers that duplicate the same capability

### Rationale

This reduces “three valid styles” to “one primary style and adapters”. It improves:
- documentation clarity
- AI generation consistency
- searchability
- migration planning

---

## 6.2 Context-aware handlers

### Decision

Retain context-aware handlers as a convenience layer, not as a co-equal route registration family.

### Proposal

Replace direct emphasis on `GetCtx/PostCtx/...` with one of these:

#### Option A: adapter helper

```go
app.Get("/health", contract.Adapt(func(ctx *contract.Ctx) {
    ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}))
```

#### Option B: utility wrapper

```go
app.Get("/health", core.ContextHandler(func(ctx *contract.Ctx) {
    ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}))
```

Either option is better than maintaining a full parallel `GetCtx/PostCtx/...` family as a first-class style.

### Rationale

This preserves ergonomic context-based handling without turning it into a second canonical app API.

---

## 6.3 `Handle` and `HandleFunc`

### Decision

The public route registration surface should not expose multiple near-equivalent generic entrypoints unless each serves a distinct role.

### Proposal

Keep:
- `Handle(pattern string, handler http.Handler)`

Review whether `HandleFunc` is still necessary if `Get/Post/...` already accept `http.HandlerFunc`.

If `HandleFunc` adds no unique clarity, deprecate it from the primary docs.

---

## 6.4 `router` package contraction

### Decision

`router` must become a pure routing package.

### Keep in `router`

- `Router`
- `Group`
- route registration internals
- path matching and params
- route freeze or immutable route table behavior
- static route mount support if truly part of serving behavior

### Migrate out of `router`

- repository abstractions
- resource controller abstractions
- JSON writer utilities
- validator composition helpers
- batch or generic CRUD scaffolding

### Target destinations

- `contract` — response writing, request decoding, error model
- `resource` or `rest` — controller-style helpers
- `validate` — validation helpers
- `store` / `persistence` — repository abstractions

### Rationale

This restores package meaning and makes search + code generation far more reliable.

---

## 6.5 Error contract contraction

### Decision

All framework-level structured HTTP errors must flow through one contract package and one model family.

### Proposal

Introduce or standardize:

```go
type Error struct {
    Code     string
    Message  string
    Status   int
    Category string
    Details  map[string]any
}
```

Required write path:

```go
contract.WriteError(w, r, err)
```

Recovery, debug, auth, rate limits, validation failures, and body-size failures must converge on this path.

### Rationale

If errors are not contracted, the repo will keep drifting into multiple JSON shapes.

---

## 6.6 Response contract contraction

### Decision

All standard JSON success responses should converge to one response helper family.

### Proposal

Canonical path:

```go
contract.WriteResponse(w, r, status, payload)
```

Optional context wrapper:

```go
ctx.Response(status, payload)
```

But both must terminate in the same contract implementation.

### Rationale

This keeps success and error writing aligned and easier for agents to follow.

---

## 6.7 Bind middleware de-emphasis

### Decision

`middleware/bind` remains available but becomes explicitly non-canonical for common application code.

### Proposal

Update docs to say:
- use explicit decode by default
- use middleware bind only when the surrounding architecture already requires request enrichment through middleware
- never use middleware bind in introductory examples or canonical examples

### Rationale

Middleware-based DTO flow is compact, but it is less legible for code agents and for human review.

---

## 6.8 Core vs extension visibility

### Decision

The repository must visually separate:
- core runtime APIs
- compatibility APIs
- experimental APIs
- extension/capability APIs

### Proposal

Adopt one of the following visibility rules:
- annotate docs with status banners
- move experimental/extension capabilities under obvious namespaces
- keep README first screen focused only on the canonical core

### Rationale

A wide first screen trains both humans and agents to treat everything as equally core.

---

## 7. API Migration Matrix

This matrix links current surface to the target state.

| Current API / Pattern | Target Status | Action |
|---|---|---|
| `app.Get/Post/...` with `http.HandlerFunc` | Canonical | Keep and document first |
| `app.GetCtx/PostCtx/...` | Compatibility / adapter | De-emphasize, migrate docs away |
| `app.GetHandler/PostHandler/...` | Deprecated | Remove from canonical docs, plan removal |
| `app.Handle(pattern, handler)` | Canonical advanced entry | Keep |
| `app.HandleFunc(...)` | Review / likely de-emphasize | Keep only if distinctly useful |
| `router.JSONWriter` | Deprecated / migrate | Move to `contract` or remove |
| resource/controller/repository types under `router` | Migrate out | Create proper package homes |
| middleware DTO injection via `bind` | Non-canonical | Keep as optional package |
| multiple JSON response writer families | Contract | Converge on `contract.WriteResponse` |
| multiple error JSON shapes | Contract | Converge on `contract.WriteError` |

---

## 8. Proposed Package Boundary After Contraction

### 8.1 Core runtime boundary

#### `core`
Owns:
- app construction
- options
- boot/shutdown lifecycle
- component registration orchestration
- route registration facade
- middleware registry integration

Must not own:
- business controller scaffolding
- persistence abstractions
- validation frameworks
- non-core capability APIs as direct convenience surface unless strictly necessary

#### `router`
Owns:
- path tree / matching
- groups
- params
- route freezing
- route mount mechanics

Must not own:
- repository patterns
- CRUD/controller base classes
- JSON/error writer APIs
- validation composition

#### `contract`
Owns:
- request context wrapper if retained
- request decode helpers
- response helpers
- framework HTTP error model
- request/response trace metadata contract

#### `middleware`
Owns:
- middleware interfaces
- registry
- core middleware implementations

### 8.2 Extension boundary

Extension packages must not leak back into canonical core docs unless the example is explicitly extension-specific.

Examples:
- `components/ops`
- `websocket`
- `webhook`
- `tenant`
- `scheduler`
- `frontend`
- `ai`

---

## 9. Phased Execution Plan

## Phase 1 — Documentation contraction

Deliverables:
- update README first screen to show only canonical core path
- update examples so “hello world”, CRUD, auth, JSON API all use one route/handler style
- add package status labels: canonical / compatibility / experimental

No breaking API changes required.

## Phase 2 — Soft deprecation

Deliverables:
- mark duplicate route registration helpers as deprecated in docs and GoDoc comments
- mark non-canonical packages/examples clearly
- introduce adapters for context handlers so migration remains easy

Still no major removals.

## Phase 3 — Package extraction

Deliverables:
- move non-routing types out of `router`
- move response writing into `contract` consistently
- ensure deprecated names keep forwarding until v1 or v1.x according to policy

## Phase 4 — v1 surface freeze

Deliverables:
- publish final canonical surface
- freeze the minimal public core API
- leave compatibility shims only where justified by migration cost

---

## 10. Suggested PR Queue

To make this executable, here is a recommended PR sequence.

### PR 1 — Canonical docs sweep

- update README examples to use one handler style
- update examples/hello, examples/api, examples/crud to canonical style
- label compatibility APIs in docs

### PR 2 — Context handler adapter

- introduce a single adapter for `CtxHandlerFunc`
- stop promoting `GetCtx/PostCtx/...` in docs

### PR 3 — Contract error unification

- define one framework error model
- route recovery/debug/auth/rate-limit responses through one writer path

### PR 4 — Response writer unification

- standardize `WriteResponse`
- remove or deprecate duplicate JSON writing helpers

### PR 5 — `router` package extraction

- move `JSONWriter`
- move repository/controller abstractions
- move validators out
- preserve forwarding aliases only if migration cost is high

### PR 6 — Public API deprecation comments

- add GoDoc deprecation notices
- add migration references to new package homes

### PR 7 — v1 freeze checklist

- verify canonical example set
- verify public package boundaries
- verify no new duplicate route registration API was added

---

## 11. Rules for AI Agents and Contributors During Contraction

Until the contraction is complete, contributors and code agents must follow these rules:

1. Do not introduce any new route registration family.
2. Do not add new helper names that duplicate existing response or error behavior.
3. Do not add non-routing abstractions to `router`.
4. Do not add middleware-based DTO injection to canonical examples.
5. When touching handler examples, migrate toward the canonical style instead of preserving mixed styles.
6. When adding new extension capabilities, keep them out of the core README path unless they are essential to bootstrapping an HTTP app.
7. Prefer explicit decode and explicit response writing in examples, tests, and generated code.

---

## 12. Acceptance Criteria

This proposal is considered implemented when all of the following are true:

- A new contributor can identify the canonical route/handler style from the first screen of the README.
- The default examples do not mix handler registration styles.
- `router` no longer exposes repository/controller/JSON-writer style responsibilities.
- Framework HTTP errors flow through one error model and one write path.
- Success responses flow through one contract writer family.
- Compatibility APIs are clearly labeled and absent from primary examples.
- The public core API can be explained in a small diagram without mentioning optional capability packs.

---

## 13. Non-Goals

This proposal does not require:
- removing useful extension packages
- banning context-aware handlers entirely
- removing components as an architectural mechanism
- forcing all users to rewrite existing code immediately

The point is contraction of the **primary mental model**, not destruction of compatibility.

---

## 14. Final Recommendation

For v1, Plumego should present itself as:

> A standard-library-first Go web runtime with a small canonical core, explicit request/response flow, and optional capability packs.

That means:
- one canonical handler style
- one canonical decode style
- one canonical response/error path
- pure package boundaries
- adapters for legacy or convenience paths
- strong separation between core and extensions

The smallest stable public surface will produce the best long-term outcomes for:
- AI-assisted development
- human review
- documentation quality
- ecosystem predictability
- future maintenance cost
