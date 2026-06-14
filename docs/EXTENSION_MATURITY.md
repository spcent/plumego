# Extension Maturity Levels

This document defines the maturity states for Plumego's optional `x/*` extension families. See also:
- [`STABILITY.md`](../STABILITY.md) — full compatibility policy
- [`docs/concepts/extension-maturity.md`](./concepts/extension-maturity.md) — technical dashboard
- [`docs/reference/extension-stability-policy.md`](./reference/extension-stability-policy.md) — policy details

---

## At a Glance

| Maturity | API Stability | Production Ready | Risk | Use Case |
|---|---|---|---|---|
| **Experimental** | May change in any minor | No | High | Pilot projects, R&D, feature exploration |
| **Beta** | Stable across release refs; breaking changes with deprecation notice | Yes, with caveats | Medium | Production when edge cases are acceptable |
| **GA** | Frozen like stable roots; breaking changes require major bump | Yes | Low | Any production system |

---

## Experimental

**API is still evolving.** Backwards compatibility is not guaranteed.

### Characteristics

- Exported symbols may be renamed, removed, or have signatures changed in any minor version.
- No deprecation notice is required before breaking changes.
- Behavior may change to fix design issues.
- Suitable for proof-of-concept and integration testing.

### When to Use

- You are evaluating the feature for your project.
- The functionality is non-critical or isolated.
- Your team owns the migration risk and can absorb API churn.
- You want early feedback on feature design.

### When Not to Use

- Critical path code in production services.
- Shared libraries depended on by multiple teams.
- Long-lived services expecting zero churn.
- When a beta or GA alternative exists.

### Current Experimental Packages

- `x/ai` (root family; subpackages `x/ai/provider`, `x/ai/session`, `x/ai/streaming`, `x/ai/tool` are beta surfaces)
- `x/data` (root family; subpackages `x/data/file` and `x/data/idempotency` are beta surfaces)
- `x/fileapi`
- `x/gateway/discovery` (subordinate)
- `x/observability/devtools` (subordinate)
- `x/observability/ops` (subordinate)
- `x/openapi`
- `x/resilience`
- `x/rpc`
- `x/validate`
- Remaining `x/data/`, `x/gateway/ipc`, `x/messaging/mq`, `x/messaging/pubsub`, `x/messaging/scheduler`, `x/messaging/webhook` subpackages

### Promotion Path

Experimental packages are promoted to beta when they meet these criteria:

1. No exported API changes across two consecutive release references.
2. Complete beta evidence document recorded in `docs/evidence/extension/`.
3. Owner sign-off.
4. Test coverage ≥70% at stable tier.
5. Complete module primer in `docs/modules/x/<package>/README.md`.

---

## Beta

**API is stable across cited release references.** Breaking changes are rare and always include a deprecation notice.

### Characteristics

- API is guaranteed stable across the two release references cited in the beta evidence document.
- Breaking changes require a deprecation notice in the PR introducing the change.
- Deprecated symbols may be removed in a subsequent minor version after the deprecation PR.
- Suitable for production adoption with the understanding that edge cases may be rough.
- Each beta package has an evidence document demonstrating API stability.

### When to Use

- Production systems where edge cases are acceptable.
- Your team can manage minor breaking changes (with deprecation notice and a clear migration path).
- The feature is essential to your product and no stable alternative exists.
- You are prepared for occasional API adjustments.

### When Not to Use

- Ultra-stable systems where any API change is unacceptable (use stable roots instead).
- Services where migration overhead is prohibitive.
- Short-lived projects where churn is high-cost.

### Current Beta Packages

- `x/frontend` — Static asset serving and embedded frontend hosting.
- `x/gateway` — Reverse proxy, rewrite, load balancing, and edge transport.
- `x/messaging` — App-facing messaging service (queue, pubsub, webhooks).
- `x/observability` — Observability exporters, tracers, and collectors.
- `x/rest` — REST resource controller and CRUD route conventions.
- `x/tenant` — Multi-tenant resolution, isolation, quota, and session management.
- `x/websocket` — WebSocket hub and explicit route registration.

### Evidence and Promotion

Each beta package has a corresponding evidence document in `docs/evidence/extension/` demonstrating:
- No API changes across two release references.
- Owner sign-off.
- Test coverage and validation results.

Example: [`docs/evidence/extension/x-rest.md`](./evidence/extension/x-rest.md)

### Promotion Path

Beta packages are promoted to GA when they meet these criteria:

1. Two full production release cycles with zero exported API changes.
2. 100% coverage at stable tier.
3. Boundary review sign-off.
4. Movement to a stable root package (from `x/*` to top-level).
5. Updated evidence document in `docs/evidence/stable-api/`.

---

## General Availability (GA)

**API is frozen.** Breaking changes require a major version bump.

### Characteristics

- API carries the same guarantee as stable root packages.
- Exported symbols, signatures, and behavior are frozen within a major version.
- Breaking changes require a major version bump (e.g., `v1.x` → `v2.0`).
- Production-ready with zero API churn expected within a major version.

### When to Use

- Any production system requiring zero API churn.
- Shared libraries and frameworks.
- Long-lived applications with minimal maintenance overhead.
- Ultra-stable systems.

### When Not to Use

- None. GA is the standard for all production systems.

### Current GA Packages

- **Stable roots only:** `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`.

GA packages are promoted from beta only after demonstrating:
- Two production release cycles with zero exported API changes.
- Full test coverage.
- Boundary review and standards alignment.

---

## How to Check Package Maturity

### Quick Check

1. Look for the package in the current section above.
2. Check [`docs/concepts/extension-maturity.md`](./concepts/extension-maturity.md) for the maturity dashboard.
3. Consult the module primer in `docs/modules/x/<package>/README.md`.

### Detailed Check

1. Read the module's `module.yaml` file (each module has a `status` field).
2. For beta packages, review the evidence document in `docs/evidence/extension/`.
3. Check the `STABILITY.md` section for the package's release history.

### Run the Drift Check

Verify maturity status against module manifests:

```bash
go run ./internal/checks/extension-maturity
```

---

## Migration Between Maturity Levels

### Within Experimental

No guarantee. API may change in any minor version. Plan for frequent updates or maintain your own fork.

### Experimental → Beta

Follow the evidence requirements in `docs/evidence/extension/`. When promoted:

1. Update your code if APIs changed (usually none, but possible).
2. You gain a deprecation-notice guarantee for breaking changes.
3. Check the beta evidence document for any caveat notes.

### Beta → GA

When promoted (rare, requires two production cycles):

1. API is now frozen like stable roots.
2. No changes needed in your code (API is already stable).
3. You gain a major-version-bump guarantee for any breaking changes.

### Experimental Direct to GA

Not possible. Packages must go through beta first to prove API stability.

---

## Support and Questions

- **"Is package X production-ready?"** → Check this document and the module primer in `docs/modules/x/<package>/README.md`.
- **"What's the risk of adopting package Y?"** → See the maturity level section above.
- **"When will package Z be promoted to beta/GA?"** → Check the evidence document or roadmap in `docs/release/roadmap.md`.
- **"Can I use experimental package W?"** → Yes, for pilots and R&D; not for production unless your team owns the risk.

Open an issue on [GitHub](https://github.com/spcent/plumego) for questions about specific packages.

---

## See Also

- [`STABILITY.md`](../STABILITY.md) — Full compatibility and promotion policy.
- [`COMPATIBILITY.md`](../COMPATIBILITY.md) — How Plumego integrates with stdlib and ecosystem.
- [`docs/concepts/extension-maturity.md`](./concepts/extension-maturity.md) — Technical maturity dashboard.
- [`docs/reference/extension-stability-policy.md`](./reference/extension-stability-policy.md) — SemVer and breaking-change policy.
- [`docs/modules/`](./modules/) — Package-specific primers.
