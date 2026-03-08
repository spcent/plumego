# Plumego Deprecation Plan v0 to v1

Status: Draft v0.1  
Audience: maintainers, contributors, adopters  
Goal: reduce API width, clarify package boundaries, and define the path to a stable v1 core

---

## 1. Purpose

This document defines the migration path from the current Plumego public surface to a **smaller, more stable v1 core**.

The repository already presents `core` as the stable primary entrypoint, while the package index shows multiple parallel handler registration styles and a broad capability surface. The goal of this plan is not to remove useful features, but to clearly separate:

- stable core runtime;
- optional capability packs;
- legacy compatibility layers;
- deprecated APIs that should stop appearing in new code.

---

## 2. Goals for v1

By v1, Plumego should provide:

- one canonical handler registration style;
- one canonical request/response/error style;
- a `router` package focused on routing only;
- explicit compatibility layers for legacy convenience APIs;
- clear package stability markers;
- a narrow `core.App` surface.

---

## 3. Deprecation Principles

### 3.1 Compatibility first, but not ambiguity forever

v0 may keep adapters. v1 should not keep multiple equally recommended ways to do the same thing.

### 3.2 Soft deprecate before remove

Every removal should pass through three steps:

1. documentation demotion;
2. code deprecation markers and migration notes;
3. removal in the next major boundary only.

### 3.3 Keep runtime behavior stable during transition

Where possible, deprecated APIs should delegate to the canonical implementation internally.

### 3.4 Prefer moves over silent rewrites

If a type belongs in another package, move or re-home it explicitly and document the replacement.

---

## 4. Versioning Stages

## Stage A — current v0.x: freeze expansion

Immediate rule:

- do not add new parallel public APIs for route registration;
- do not add more non-routing concerns into `router`;
- do not add optional platform features directly onto `core.App` without an explicit exception review.

## Stage B — late v0.x: document canonical path

Required deliverables:

- publish `AI_FRIENDLY_CORE_GUIDE.md`;
- publish a canonical quick start using one handler style;
- mark deprecated APIs in docs and code;
- label optional/experimental packages clearly.

## Stage C — v1.0 release candidate: enforce package boundaries

Required deliverables:

- route all canonical examples through the v1 core path only;
- complete moves out of `router` for non-routing types;
- ensure deprecated APIs are wrappers only, not separate implementations.

## Stage D — v1.0

At v1:

- the canonical surface is frozen;
- deprecated v0 compatibility APIs may still exist if cheap to maintain, but must no longer appear in first-party examples or guides;
- any further removals wait for v2 unless marked experimental from the beginning.

---

## 5. API Review by Category

## 5.1 Handler registration APIs

### Current state

`core.App` exposes multiple route registration families, including:

- `Get/Post/Put/Delete/...` using `http.HandlerFunc`;
- `GetCtx/PostCtx/PutCtx/DeleteCtx/...` using `contract.CtxHandlerFunc`;
- `GetHandler/PostHandler/PutHandler/DeleteHandler/...` using `router.Handler`.

### Problem

These are parallel primary APIs.

### v1 decision

#### Canonical

Keep verb registration based on `http.HandlerFunc` as the canonical external path.

#### Compatibility

- keep `*Ctx` registrations during v0.x as compatibility adapters;
- keep `*Handler` registrations only if they are thin wrappers over the same underlying implementation.

#### Documentation status

- `Get/Post/...`: canonical;
- `GetCtx/PostCtx/...`: compatibility, not for new docs;
- `GetHandler/PostHandler/...`: compatibility or advanced adapter, not for new docs.

### Required tasks

1. mark `*Ctx` and `*Handler` families as non-canonical in docs;
2. route all examples to `http.HandlerFunc` style;
3. ensure all registration forms call the same internal core registration path.

### Optional future removal

Actual removal of non-canonical forms can wait until v2 if the maintenance cost is low.

---

## 5.2 `core.App` capability methods

### Current state

`core.App` includes helpers such as websocket, webhook, pubsub, auth, and rate-limit related methods, including `ConfigureWebSocketWithOptions`, `ConfigureWebhookIn`, `ConfigureWebhookOut`, `EnableAuth`, and `EnableRateLimit`.

### Problem

This broadens the app surface and makes `App` feel like a product platform instead of a minimal runtime.

### v1 decision

#### Keep in v1 only if

- the method is necessary for most HTTP applications; and
- it represents runtime infrastructure rather than optional capability.

#### Preferred direction

Move optional capability setup behind component packages or extension packages, while keeping `App` focused on runtime primitives.

### Required tasks

1. review each non-routing `App` method against the admission checklist;
2. move optional features to extension/component packages where practical;
3. retain only the minimal necessary convenience methods on `App`.

### Suggested status map

- auth/rate-limit bootstrap helpers: review;
- websocket/webhook/pubsub configuration: demote from core-facing docs; consider package-level builders or components.

---

## 5.3 `router` package surface

### Current state

The `router` package includes routing primitives and also non-routing types such as `BaseRepository`, `BaseResourceController`, and `JSONWriter`. `JSONWriter` is already documented as deprecated in favor of contract-level response writing.

### Problem

Package responsibility is too broad.

### v1 decision

`router` becomes routing-only.

### Required tasks

1. move `JSONWriter` replacement guidance fully to `contract`;
2. move base controller/resource helpers to a new package such as `rest`, `resource`, or `scaffold`;
3. move repository helpers out of `router` into a persistence-oriented package;
4. keep route params, groups, and matching in `router` only.

### Deprecation status

- `router.JSONWriter`: already deprecated, continue migration;
- `router.BaseResourceController`: legacy helper, not part of v1 core;
- `router.BaseRepository`: legacy helper, not part of v1 core.

### Compatibility policy

These types may remain in v1 as legacy helpers if low-cost, but:

- they must be excluded from canonical docs;
- their package placement should be treated as historical, not aspirational.

---

## 5.4 Bind middleware style

### Current state

`middleware/bind.BindJSON[T]` parses JSON and validates payloads as a compatibility middleware; canonical style is explicit decode and validation inside handlers.

### Problem

It creates implicit request data flow.

### v1 decision

Keep bind middleware as an optional helper, not the default request decoding style.

### Required tasks

1. move explicit JSON decode examples ahead of bind examples;
2. document bind as convenience for specific pipelines;
3. avoid introducing more hidden bind flows that combine body/query/path automatically.

### Deprecation status

No removal required. This is a **documentation de-promotion**, not a feature deletion.

---

## 5.5 Error writing APIs

### Current state

There is evidence of migration toward contract-level response writers, and `router.JSONWriter` is deprecated in favor of `contract.WriteResponse/WriteError` for new code.

### Problem

If multiple packages keep defining response/error-writing styles, new code will drift.

### v1 decision

`contract` owns the canonical request/response/error writing contract.

### Required tasks

1. standardize all new examples on `contract.WriteJSON` and `contract.WriteError` style APIs;
2. ensure middleware and recovery paths use the same contract shape;
3. avoid new package-local error payload formats.

### Deprecation status

- package-local writer types: deprecate in docs;
- contract-level writer path: canonical.

---

## 6. Package Stability Labels

By the end of v0.x, every public-facing package should be labeled as one of:

- **stable core**;
- **stable extension**;
- **experimental**;
- **legacy compatibility**.

### Suggested initial labeling

- `core`: stable core;
- `router`: stable core after surface slimming;
- `middleware`: stable core/extension depending on package;
- `contract`: stable core;
- `ai`, `scheduler`, `tenant`, `frontend`, `messaging`, `pubsub`, `store`, `security`: stable extension or experimental depending on maturity. The repository currently exposes these package families at top level.

---

## 7. Migration Table

| Current API / Area | Current Status | v1 Status | Action |
|---|---|---|---|
| `App.Get/Post/...` | active | canonical | keep |
| `App.GetCtx/PostCtx/...` | active | compatibility | de-emphasize in docs, keep as adapter |
| `App.GetHandler/PostHandler/...` | active | compatibility / advanced | de-emphasize in docs, unify implementation |
| `App` optional capability helpers | active | review | move many to extension/component setup |
| `router.JSONWriter` | deprecated | legacy only | replace with `contract` writers |
| `router.BaseResourceController` | active | not core | move or mark legacy helper |
| `router.BaseRepository` | active | not core | move or mark legacy helper |
| `middleware/bind` default-style usage | active | optional convenience | de-promote in docs |
| package-local error writers | mixed | discouraged | consolidate into `contract` |

---

## 8. Documentation Migration Plan

## Phase 1

Update README and quick start so the first visible path is:

- `core.New(...)`
- `app.Use(...)`
- `app.Get(...)`
- explicit JSON decode
- `contract.WriteJSON / WriteError`

## Phase 2

Add labels in package docs:

- canonical;
- compatibility;
- deprecated;
- experimental.

## Phase 3

Reorganize examples into:

- `examples/canonical/...`
- `examples/advanced/...`
- `examples/experimental/...`

This reduces the risk that agents learn optional patterns as if they were the primary framework style.

---

## 9. Code Change Plan

### Immediate

- add deprecation comments where missing;
- stop adding new public parallel APIs;
- route old handlers through canonical internal registration.

### Next

- create new homes for non-routing helpers currently inside `router`;
- add one canonical error writer path in docs and tests;
- rewrite first-party examples to one style.

### Before v1 tag

- audit package names and move obvious boundary violations;
- ensure deprecated APIs are wrappers, not forked code paths;
- publish a v1 surface statement.

---

## 10. What Will Not Be Deprecated Yet

To keep transition cost reasonable, the following should generally stay available through v1 unless they prove expensive to maintain:

- thin compatibility adapters for alternate handler forms;
- useful extension packages that do not pollute the core path;
- optional bind middleware as long as it remains clearly non-canonical.

The point of this plan is to reduce **ambiguity**, not to delete useful code aggressively.

---

## 11. Exit Criteria for v1

Plumego is ready for a v1 core tag when all of the following are true:

1. one canonical handler style is documented and used everywhere in first-party quick starts;
2. `router` is routing-focused in both package contents and mental model;
3. `contract` owns the canonical response/error-writing path;
4. optional capabilities no longer dominate the `core.App` surface;
5. compatibility APIs are labeled clearly;
6. contributors can tell, at a glance, which APIs are core, extension, experimental, or legacy.

---

## 12. Final Direction

v1 should not try to freeze every useful package in the repository.

v1 should freeze a **small, coherent HTTP core** and let the rest of Plumego evolve around it.
