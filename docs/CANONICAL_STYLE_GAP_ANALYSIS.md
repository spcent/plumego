# Canonical Style Gap Analysis for Plumego

This document summarizes where the current repository diverges from
`docs/CANONICAL_STYLE_GUIDE.md` and proposes a prioritized optimization plan.

## Scope and Method

The analysis focused on:

- Top-level docs (`README.md`, `README_CN.md`, `docs/getting-started.md`)
- API docs (`docs/api/`, `docs/modules/`)
- Official examples under `examples/`

Primary checks:

1. Canonical handler/route style consistency (`net/http` shape first)
2. Canonical request decode style (explicit JSON decode in handlers)
3. Canonical error/success write path consistency
4. Canonical-vs-compatibility labeling discipline
5. Separation of “core learning path” vs advanced/compatibility content

## Baseline Expectations from the Canonical Guide

The style guide sets these key defaults:

- One primary route registration and handler style for onboarding docs
- Explicit decode in handlers for canonical examples
- One stable error write path
- Constructor-based dependency flow (no context service locator)
- Compatibility APIs can exist, but should not lead canonical docs
- Official docs/examples should clearly label `canonical`, `compatibility`, or `experimental`

## Current Gaps and Optimization Opportunities

### 1) Canonical and compatibility styles are still mixed in major docs

**Observed issues**

- Router API docs prominently present `*Ctx` variants alongside canonical forms in the main registration table.
- Contract request docs are largely centered on `PostCtx/GetCtx` + context binding workflows.
- Middleware docs (EN/ZH examples) still teach `GetCtx/PostCtx`, `contract.Ctx`, and middleware-first binding patterns in core walkthrough sections.

**Why this matters**

This widens the style surface in first-read docs and conflicts with the “one obvious way” onboarding principle.

**Recommended optimization**

- Keep compatibility APIs documented, but move them into dedicated compatibility sections/pages.
- Make canonical sections strictly `func(http.ResponseWriter, *http.Request)` oriented.
- Add explicit callouts such as “Compatibility API (not recommended for new code)” where legacy forms are retained.

### 2) Request binding docs still default to auto/middleware/context-centric flows

**Observed issues**

- `docs/modules/contract/request.md` emphasizes `ctx.Bind`, `BindJSON`, `BindForm`, and `BindQuery` via `PostCtx` and `GetCtx`.
- `examples/bind-example/main.go` demonstrates middleware-first DTO binding (`bind.BindJSON` + `bind.FromRequest`) as the main flow.

**Why this matters**

Canonical guidance prefers explicit source-visible decoding in handlers for default teaching paths.

**Recommended optimization**

- Reframe bind middleware as an advanced or compatibility pattern.
- Add canonical transport examples that use explicit `json.NewDecoder(r.Body).Decode(&req)` first.
- Keep bind middleware docs, but clearly tag as optional and non-canonical for general CRUD onboarding.

### 3) Bootstrap and response guidance remains partially split

**Observed issues**

- README quick start centers `app.Boot()` and option-heavy startup, while another section uses `http.ListenAndServe`.
- README highlights `contract.WriteResponse` / `Ctx.Response` in “Contract Helpers,” which can steer examples toward context-heavy response patterns.

**Why this matters**

Canonical onboarding should present one stable startup and response writing path before alternatives.

**Recommended optimization**

- Establish one canonical startup snippet (prefer stdlib server compatibility-first path).
- Keep `Boot()` and context helpers in an “advanced runtime APIs” or “compatibility” section.
- Pick one canonical success/error write pair for docs and use it consistently across top-level guides.

### 4) Example portfolio has high compatibility-style density in flagship demos

**Observed issues**

- `examples/multi-tenant-saas/main.go` heavily uses `GetCtx/PostCtx/PutCtx/DeleteCtx`.
- `examples/crud-demo/main_simple.go` route registration uses `*Ctx` methods by default.
- `examples/sms-gateway/main.go` uses `PostCtx` in externally visible API endpoints.

**Why this matters**

Large examples influence contributor patterns more than API docs. If these remain context-first, canonical adoption slows.

**Recommended optimization**

- For each major example, provide a canonical route surface variant.
- If retaining compatibility-style examples, tag file headers and README sections as `compatibility`.
- Add a “canonical reference path” index that points newcomers to `examples/reference/main.go` first.

### 5) Labeling discipline is inconsistent across docs/examples

**Observed issues**

- The style guide requires clear labels (`canonical`, `compatibility`, `experimental`), but most docs/examples are not explicitly labeled.
- `examples/reference/main.go` is labeled canonical, but many neighboring docs are unlabeled while still teaching compatibility APIs.

**Why this matters**

Without labels, users and AI agents cannot quickly distinguish preferred long-term patterns from legacy/advanced forms.

**Recommended optimization**

- Add a front-matter or first-heading tag convention in docs/examples.
- Enforce labeling via docs review checklist (or CI lint on markdown headings/tags).

## Prioritized Execution Plan

### Priority 0 (Immediate, low-risk docs-only)

1. Update top-level README quick-start to one canonical style path.
2. Add explicit compatibility labels to docs that teach `*Ctx`/binding middleware patterns.
3. Add “Canonical Learning Path” table linking: `docs/getting-started.md` → `examples/reference/main.go`.

### Priority 1 (High impact onboarding alignment)

1. Refactor `docs/modules/contract/request.md` into:
   - Canonical (explicit decode)
   - Compatibility (`ctx.Bind*`, `*Ctx`)
2. Refactor `docs/api/router.md` so route registration starts with canonical API only; move `*Ctx` table rows into compatibility subsection.
3. Update middleware docs to separate transport-only canonical middleware guidance from context helper compatibility recipes.

### Priority 2 (Example portfolio alignment)

1. Add canonical variants for `crud-demo`, `multi-tenant-saas`, and `sms-gateway` public route wiring.
2. Retain existing compatibility examples but rename or label accordingly.
3. Ensure each example README starts with “Style: canonical|compatibility|experimental”.

### Priority 3 (Governance and drift prevention)

1. Add a docs CI check ensuring style labels exist in official examples/docs directories.
2. Add PR checklist items mapped directly to canonical style sections.
3. Schedule incremental migration for high-traffic docs first (README, getting-started, router docs).

## Suggested Success Metrics

- Reduce `GetCtx/PostCtx/...` appearances in onboarding docs by >=80%.
- Ensure 100% of official example READMEs have style labels.
- Ensure top-level README uses one canonical handler + startup + response story.
- Keep compatibility APIs documented but never leading in first-read onboarding sections.

## Notes

- This plan does not require removing compatibility APIs.
- It focuses on documentation and example posture so that default adoption aligns with the canonical guide while preserving backward compatibility.
