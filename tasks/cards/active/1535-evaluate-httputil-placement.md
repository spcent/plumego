# Card 1535 — Evaluate BufferedResponse placement and httputil splitting

Milestone: M-027
Recipe: (none — design/architecture)
Context Package: internal
Priority: P2
State: active
Primary Module: internal/httputil
Owned Files: (evaluation only — no code changes in this card)
Depends On: 1534 (blocked until error-code consolidation is resolved)

## Goal

Evaluate whether `internal/httputil.BufferedResponseRecorder` (and related `ResponseRecorder`) should be promoted to a more accessible layer, or whether `internal/httputil` should be split into smaller, single-purpose internal packages.

## Problem

`internal/httputil` is the most widely depended-on internal package in the repository:

- **14 consumers**: 12 middleware + `x/gateway/cache`, `x/gateway/transform`, `x/messaging/webhook`
- **Dual responsibilities**: buffered response recording + HTTP utilities (ClientIP, SafeWrite, CopyHeaders)
- **Implicit stable contract**: middleware treat `BufferedResponseRecorder` as infrastructure, but it lives in `internal`

This centralization means:
- Any change to `BufferedResponseRecorder` affects 14 modules
- New response-recorder variants must be added to the same bucket
- `internal/httputil` is large enough (~280 lines) that its consumers import more than they need

## Scope of Evaluation

Three design paths (pick one or reject all):

### Option A: Promote `BufferedResponse` to `contract/buffer`
- Move `BufferedResponseRecorder` + `NewBufferedResponse` to `contract/buffer`
- Keep `httputil` for utilities only (ClientIP, SafeWrite, etc.)
- Pros: response recording becomes a first-class transport primitive; no `internal` gate
- Cons: expands `contract` further; may violate `contract` = "transport primitives only" boundary

### Option B: Split `internal/httputil` into `internal/buffer` + `internal/netutil`
- `internal/buffer`: `BufferedResponseRecorder`, `ResponseRecorder`
- `internal/netutil`: `ClientIP`, `SafeWrite`, `CopyHeaders`, `AddVary`
- Pros: each package has one role; callers import exactly what they need
- Cons: more files; `dependency-rules.yaml` more complex

### Option C: Keep `internal/httputil` as-is, clarify docs
- Add package doc explaining what belongs here and what does not
- Define "new response recorder variants → here; new HTTP client utilities → no"
- Pros: zero churn; no breaking changes
- Cons: does not solve centralization

## Non-goals

- No code changes in this card — evaluation and ADR only
- No re-export changes until 1534 is resolved
- No decision is final until documented in `docs/concepts/`

## Deliverable

A **decision record** in one of the following locations:
- `docs/concepts/internal-httputil-scope.md` (Option C)
- `specs/change-recipes/move-buffered-response.yaml` (Option A or B)

## Acceptance Criteria

- All three options evaluated with trade-offs
- Decision recorded and linked from `internal/httputil/doc.go`
- If Option A/B chosen, dependency-rules.yaml impact documented

## Done Definition

- [ ] Evaluation document written
- [ ] Decision recorded with rationale
- [ ] `internal/httputil/doc.go` links to decision
- [ ] Stakeholders (AGENTS.md authority order) consulted if scope expands
