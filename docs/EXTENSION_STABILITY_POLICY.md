# Extension Stability Policy

This document defines the criteria for advancing an `x/*` extension module from
`experimental` to `beta` (stable-candidate), and from `beta` to `ga`.

It does **not** override the stable-root compatibility promise defined in
`docs/DEPRECATION.md`. Stable roots (`core`, `router`, `contract`, `middleware`,
`security`, `store`, `health`, `log`, `metrics`) follow a separate, stronger
policy.

---

## Status Ladder

The `status` field in each module's `module.yaml` tracks position on this ladder:

| Status | Meaning |
|---|---|
| `experimental` | API shape may change; no compatibility expectation |
| `beta` | API shape is stable within the current major version; breaking changes require deprecation notice |
| `ga` | Full v1 compatibility promise; follows `docs/DEPRECATION.md` |

All `x/*` modules start as `experimental`. Promotion is explicit and
requires meeting the criteria below.

Promotion evidence is tracked in `specs/extension-beta-evidence.yaml`. The
policy below defines the criteria; the evidence file records candidate modules,
release refs, exported API snapshot refs, owner sign-off state, and current
blockers. A module remains `experimental` until both the evidence file and the
module manifest are updated in a promotion card.

---

## Criteria for `experimental` → `beta`

An extension may be proposed as `beta` when all of the following are true:

1. **Stable public API surface.** No exported symbol changes have been needed
   for at least two consecutive minor releases. All public types use constructor
   injection rather than mutable fields or global registration.

2. **Boundary compliance.** The module passes `go run ./internal/checks/dependency-rules`
   with no violations. It does not import stable roots in ways that would force
   stable-root changes to accommodate it.

3. **Test coverage.** The module has unit tests for every documented public
   behavior path, including negative paths (errors, empty inputs, context
   cancellation). The test suite runs cleanly with `go test -race ./...`.

4. **Module manifest.** The `module.yaml` is complete and schema-valid
   (`go run ./internal/checks/module-manifests`). `responsibilities`,
   `non_goals`, `review_checklist`, and `agent_hints` accurately describe the
   current implementation.

5. **Module primer.** `docs/modules/<family>/README.md` documents all public
   entrypoints, boundary rules, and a validation command. It is consistent with
   the current API surface (not aspirational).

6. **No known regressions.** No open regression reports against the module's
   documented behavior.

7. **Owner sign-off.** The module owner listed in `module.yaml` confirms the
   criteria are met.

---

## Criteria for `beta` → `ga`

In addition to maintaining all `beta` criteria, a `beta` module must:

1. **Production usage evidence.** At least one production deployment (internal
   or external) is documented or known to the owner.

2. **Two-release stability.** The `beta` status has been held for at least two
   consecutive minor releases with no breaking changes.

3. **Deprecation pathway.** Any symbols previously deprecated while in
   `experimental` or `beta` have been removed or have a documented removal
   timeline.

4. **GA compatibility claim reviewed.** The module owner and a stable-root
   reviewer have confirmed the public surface is ready for the full
   `docs/DEPRECATION.md` promise.

---

## Promotion Process

1. Open a task card in `tasks/cards/active/` referencing this policy.
2. Update `specs/extension-beta-evidence.yaml` with the required release refs,
   exported API snapshot refs, blocker state, and owner sign-off.
3. Generate or compare exported API snapshots with
   `go run ./internal/checks/extension-api-snapshot`.
4. Update the `status` field in the module's `module.yaml`.
5. Update `docs/modules/<family>/README.md` to reflect the new status.
6. Update `docs/ROADMAP.md` to record the promotion.
7. The CI-equivalent release gate must pass before merging:
   ```
   make gates
   ```

---

## Current Evaluation

The following extensions are the most likely candidates for `beta` based on API
maturity and test coverage. This is a starting assessment, not a commitment.
The canonical blocker ledger for these candidates is
`specs/extension-beta-evidence.yaml`; supporting evidence notes live under
`docs/extension-evidence/`.

| Module | Candidate for | Current status / remaining action |
|---|---|---|
| `x/rest` | `beta` | CRUD negative-path tests, boundary docs, and primer coverage are in place; promotion is blocked until release-history evidence proves two consecutive minor releases without exported API changes |
| `x/websocket` | `beta` | hub lifecycle negative-path tests, boundary docs, and primer coverage are in place; promotion is blocked until release-history evidence proves two consecutive minor releases without exported API changes |
| `x/tenant` | `beta` | Resolution, policy, quota, rate-limit, session, and tenant-aware store coverage are substantially complete; promotion is blocked until release-history evidence proves two consecutive minor releases without exported API changes |
| `x/observability` | `beta` | Primer, exporter, collector, tracer, and tested-subpackage coverage are in place; promotion is blocked until release-history evidence proves two consecutive minor releases without exported API changes |
| `x/gateway` | `beta` | Gateway, backend, circuit-breaker, balancer, proxy, rewrite, transform, cache, and protocol middleware tests are in place; promotion is blocked until release-history evidence proves two consecutive minor releases without exported API changes |

Extensions not yet evaluated or with clear open work:

- `x/ai` — stable-tier subpackages (`provider`, `session`, `streaming`, `tool`) may
  be evaluated individually when orchestration remains experimental; do not
  promote the whole family while orchestration, semantic cache, marketplace,
  distributed execution, and resilience wrappers still carry experimental
  contracts
- `x/discovery` — new Kubernetes/etcd backends need two-release observation period
- `x/data` — sharding, rw, idempotency, kvengine, and file metadata coverage are
  improving; evaluate topology-heavy features as a unit
- `x/scheduler`, `x/webhook`, `x/messaging`, `x/mq`, `x/pubsub` — subordinate
  family members; evaluate after canonical root stabilizes

---

## Non-Goals

- Do not promote `x/*` packages to `ga` without this process.
- Do not weaken the stable-root promise to accommodate extension promotion.
- Do not let `beta` status become a permanent holding pattern; set a target
  release for `ga` at promotion time or record the blocker explicitly.

---

## See Also

- `docs/DEPRECATION.md` — full v1 compatibility policy for stable roots
- `specs/module-manifest.schema.yaml` — `status` enum and manifest rules
- `docs/ROADMAP.md` — current phase status
- `AGENTS.md` — quality gates and workflow rules
