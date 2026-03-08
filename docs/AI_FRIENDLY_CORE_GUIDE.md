# Plumego AI Friendly Core Guide

Status: Draft v0.1  
Audience: framework maintainers, contributors, code agents  
Scope: `core`, `router`, `middleware`, `contract`, canonical examples

---

## 1. Purpose

This document defines the design constraints for keeping Plumego's **core runtime small, predictable, and AI-friendly**.

“AI-friendly” in Plumego means:

- code generation should converge on one canonical style;
- agents should be able to modify one handler or one middleware without understanding the whole repository;
- package names should match responsibilities closely;
- public APIs should stay small enough to fit into limited context windows;
- framework behavior should be explicit rather than hidden in convenience layers.

Plumego is already standard-library-first and zero-dependency, which is a strong foundation. The next step is to keep the **core narrow** while moving broader platform capabilities into optional packages. The repository currently exposes a wide package surface, including `ai`, `scheduler`, `tenant`, `frontend`, `security`, `store`, and more, while the README positions `core` as the stable primary entrypoint.

---

## 2. Core Positioning

### 2.1 What Plumego Core Is

Plumego core is the smallest stable runtime required to build HTTP services on top of the Go standard library:

- application bootstrap;
- route registration;
- middleware composition;
- request/response baseline abstractions;
- error writing contract;
- testability with `net/http` and `httptest`.

### 2.2 What Plumego Core Is Not

Core is **not** the place for:

- platform integrations;
- task queues;
- schedulers;
- tenant management;
- AI gateways or agent workflows;
- repository/controller scaffolding;
- heavy binding and validation pipelines;
- websocket/webhook/productivity helpers directly on the main app surface.

The current repository includes many of these capabilities, which is fine at the repository level; this guide only says they should not widen the canonical core mental model.

---

## 3. Non-Negotiable Design Rules

## Rule 1: Standard library compatibility comes first

All primary abstractions must stay close to:

- `http.Handler`
- `http.HandlerFunc`
- `http.Request`
- `http.ResponseWriter`
- `func(http.Handler) http.Handler`

Plumego already follows this well in its middleware model and app serving model. The core must continue to prefer adapters over custom worlds.

### Required

- every route must be expressible as an `http.Handler`;
- every middleware must remain compatible with standard `http.Handler` composition;
- tests must work naturally with `httptest`.

### Avoid

- framework-only execution pipelines;
- lifecycle-only request abstractions that cannot be reduced to standard library handlers.

---

## Rule 2: One canonical handler style

`core.App` currently exposes multiple parallel registration forms, including `Get`, `GetCtx`, and `GetHandler`, with the same pattern repeated for other HTTP verbs. 

This increases public API width and causes style drift.

### Canonical decision

Plumego must publish **one primary handler style** for new code.

#### Recommended primary style

Use standard library handlers as the canonical external API:

```go
app.Get("/users", func(w http.ResponseWriter, r *http.Request) {
    // ...
})
```

#### Secondary style

`Ctx`-based handlers may exist as adapters or opt-in ergonomics, but they must not be treated as a peer primary style in documentation, examples, or new top-level API growth.

### Required

- all quick starts use one handler form only;
- all canonical examples use one handler form only;
- package docs explicitly label other forms as adapters or legacy compatibility layers.

### Avoid

- adding more parallel registration families;
- documenting multiple “equally recommended” handler styles.

---

## Rule 3: Router must be a router

The `router` package currently contains route concerns alongside repository, resource-controller, validator, and response-writer style utilities such as `BaseRepository`, `BaseResourceController`, and `JSONWriter`. `JSONWriter` is already marked deprecated in favor of contract-level writers.

This is too much responsibility for one package.

### Router package should contain

- route registration;
- route matching;
- groups/prefixes;
- path params;
- route freezing;
- route-level middleware composition.

### Router package should not contain

- repository primitives;
- generic CRUD controllers;
- response rendering helpers;
- validator registries;
- persistence abstractions.

### Required

- package names must map cleanly to responsibilities;
- new router APIs must be justified as routing concerns.

### Avoid

- moving “convenient” but non-routing helpers into `router`.

---

## Rule 4: Context must stay thin

A Plumego context abstraction, if kept, must remain a thin request helper, not a service locator or framework god object.

### Allowed responsibilities

- request access;
- response writing;
- path/query/header lookup;
- tiny convenience helpers for explicit I/O.

### Disallowed responsibilities

- hidden dependency lookup;
- repository access;
- validator registries;
- scheduler or queue access;
- auth/session service location;
- arbitrary component registry reads in handlers.

### Required

- business dependencies are injected through structs/functions, not fetched from context;
- context helpers must stay obvious and side-effect-light.

---

## Rule 5: Explicit data flow over implicit bind pipelines

The bind package provides `BindJSON[T]` as a compatibility middleware that parses JSON and validates payloads, but canonical HTTP handlers should decode directly from `r.Body`.

This is useful, but it creates implicit data flow.

### Canonical request parsing order

1. explicit decode in handler;
2. thin helper wrapping explicit decode;
3. middleware-based binding only for specific cases.

### Canonical preferred style

```go
var req CreateUserRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    contract.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
    return
}
```

### Required

- official examples prefer explicit decoding;
- middleware bind is documented as optional convenience, not the default style.

### Avoid

- auto-merging path/query/header/body into one implicit DTO;
- requiring handlers to know invisible middleware state to understand their inputs.

---

## Rule 6: One error contract for the entire core

The framework already has structured error-writing directions, and some older writer helpers are deprecated in favor of contract-level response writers.

Plumego core must provide a single canonical error contract that is reused across:

- handlers;
- middleware;
- recovery;
- auth/rate-limit/body-limit failures;
- debug and production-safe output paths.

### Required shape

The exact struct can evolve, but the contract must stay unified.

Recommended baseline:

```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}
```

### Required

- one canonical `WriteError` path;
- middleware must not invent package-specific error payloads without strong reason;
- debug output may enrich, but must not replace the canonical base shape.

### Avoid

- plain-text errors in one package and structured errors in another;
- multiple public error writer families.

---

## Rule 7: Middleware stays pure and predictable

Middleware should continue to use standard wrapping style. README already describes middleware as wrapping standard `http.Handler`, which is the correct baseline.

### Required

- middleware does one horizontal concern at a time;
- request control flow remains predictable;
- middleware ordering semantics are documented and stable.

### Good middleware responsibilities

- logging;
- recovery;
- timeout;
- compression;
- body-size limits;
- auth gates;
- request ids.

### Avoid

- hidden business logic;
- middleware that silently injects large dependency graphs;
- middleware that becomes the main place where DTO or domain rules are assembled.

---

## Rule 8: `core.App` is runtime glue, not a product surface

`core.App` currently includes route registration and also a broad set of configuration helpers such as websocket, webhook, pubsub, auth, and rate-limit related methods.

This is where core can gradually become too wide.

### Required

- new capabilities should default to optional packages/components before becoming `App` methods;
- `App` should expose only the smallest stable runtime surface;
- convenience methods need a high bar for admission.

### Admission checklist for new `App` methods

A new method may be added only if it:

1. is necessary for most HTTP services;
2. cannot be expressed cleanly through existing primitives;
3. does not create a new parallel style;
4. does not pull optional capability semantics into the main surface.

### Avoid

- direct `App` expansion for every optional feature;
- promoting experimental features into permanent app methods too early.

---

## Rule 9: Canonical examples outrank broad examples

The repository can keep many examples. But the framework must publish a short, obvious, canonical path.

### Required canonical example set

The docs must maintain a single minimal sequence:

1. hello world;
2. middleware + route;
3. JSON request + JSON response;
4. error writing;
5. testing with `httptest`.

### Required

- these examples are the first thing code agents see;
- other examples are explicitly labeled advanced, optional, or experimental.

---

## Rule 10: Package taxonomy must communicate stability

The repository already has a broad package map. To keep contributor and agent behavior aligned, packages must communicate stability and intent clearly.

### Recommended taxonomy

- `core`: stable runtime entrypoint;
- `router`: routing only;
- `middleware`: HTTP cross-cutting concerns;
- `contract`: request/response/error contracts and adapters;
- `experimental/...`: unstable capability packs;
- `contrib/...`: integrations and opinionated helpers.

### Required

- deprecations are explicit in code and docs;
- experimental packages are labeled prominently;
- package moves are planned rather than ad hoc.

---

## 4. Contributor Rules

Every new core contribution must answer these questions in the PR description:

1. Does this change reduce or increase the number of public ways to do the same thing?
2. Is this a routing concern, middleware concern, or optional capability concern?
3. Can a new user understand this feature without reading three other packages?
4. Can an agent patch a handler using this API without needing whole-repo context?
5. Should this belong in `core`, or should it live in an extension package?

If the answer increases API width or hides data flow, the default decision should be **no**.

---

## 5. Canonical Coding Style for New Docs

### Route registration

Use one style only:

```go
app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
    contract.WriteJSON(w, http.StatusOK, map[string]string{"message": "pong"})
})
```

### Request decode

Use explicit decode first:

```go
var req CreateUserRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    contract.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
    return
}
```

### Error response

Use one writer path only:

```go
contract.WriteError(w, http.StatusBadRequest, "bad_request", "missing user id", nil)
```

### Middleware

Use pure wrapper style:

```go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        next.ServeHTTP(w, r)
    })
}
```

---

## 6. Core Acceptance Checklist

A change is acceptable for core only if all are true:

- it preserves standard library compatibility;
- it does not add another parallel primary API;
- it keeps package responsibility narrow;
- it does not introduce hidden request data flow by default;
- it strengthens canonical examples instead of weakening them;
- it improves, or at least does not harm, local reasoning by contributors and agents.

---

## 7. Immediate Priorities

Based on the current repository state, the next core-focused priorities are:

1. declare one canonical handler style and demote the others to adapters or compatibility paths;
2. slim `router` down to routing concerns only;
3. consolidate one framework-level error contract and writer entrypoint;
4. demote middleware-based binding from default style to optional helper;
5. split “core runtime” from “capability packs” more clearly in docs and package layout.

---

## 8. Final Principle

Plumego should be easy to understand by reading one handler, one middleware, and one route registration block.

If a contributor or an AI agent must inspect multiple packages to infer basic request flow, the core has become too wide.
