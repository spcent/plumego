# Plumego Deprecation Policy

This document defines how Plumego manages deprecation of public APIs, extension
packages, and configuration surfaces across stable releases.

---

## Scope

This policy applies to:

- **Stable library roots** (`core`, `router`, `contract`, `middleware`,
  `security`, `store`, `health`, `log`, `metrics`) — `ga` compatibility promise
  starting at v1.
- **Canonical reference app** (`reference/standard-service`) — kept aligned with
  stable-root API, not a reusable extension catalog.
- **CLI** (`cmd/plumego`) — supported command-line tool; not a Go import
  surface.

It does **not** apply to:

- `x/*` extension packages — see below for tier-specific rules.
- Internal packages (`internal/`, unexported symbols) — Implementation details;
  may change without notice.
- Test helpers inside `_test.go` files.
- Pre-release migration wrappers and compatibility shims that have no external
  compatibility promise.

Agent cleanup rule: once the last in-repository caller has moved off a
pre-release or internal compatibility wrapper, remove that wrapper in the same
change. Do not preserve dead wrappers just because they were useful during a
migration. Released stable-root public APIs follow the compatibility promise
below.

---

## Extension Package Tiers

`x/*` extension packages follow the maturity ladder defined in
`docs/EXTENSION_STABILITY_POLICY.md`. The compatibility rules differ by tier:

| Tier | Modules | Compatibility rule |
| --- | --- | --- |
| `beta` | `x/gateway`, `x/observability`, `x/rest`, `x/websocket` | Exported API frozen between minor release refs. Breaking changes require a new tagged ref and a snapshot comparison showing the diff. No silent breakage. |
| `experimental` | all other `x/*` | No compatibility freeze. API and config may change between any commits. Adopt for clear reasons, not by default. |
| `ga` | none yet | Full stable-root-equivalent compatibility promise. |

Promotion from `experimental` to `beta` requires evidence recorded in
`specs/extension-beta-evidence.yaml` — see
`docs/extension-evidence/BETA_EVIDENCE_TEMPLATE.md` and
`docs/release/PROMOTION_CARD_TEMPLATE.md` for the workflow.

---

## Compatibility Promise (v1 Stable Roots)

For every exported symbol (type, function, interface, constant, variable) in a
stable root:

- Its signature will not change in a backward-incompatible way within a major
  version.
- Packages will not be renamed or removed within a major version.
- Behavior documented in godoc will not regress silently.

A new major version (e.g., v2) is the supported mechanism for breaking
changes.

Exported API baseline snapshots are recorded under `docs/stable-api/snapshots/`.
They are review evidence for stable-root freeze work and are updated alongside
each release.

---

## Deprecation Process

### Step 1 — Mark in code

Add a `// Deprecated:` godoc comment immediately above the symbol, including:

- What replaces it (package path, function name, or approach).
- Which release the deprecation was introduced.
- The planned removal release (major version, or "TBD" if not yet scheduled).

```go
// Deprecated: Use middleware/httpmetrics.Middleware instead.
// Deprecated since: v1.0.0. Planned removal: v2.
func OldMiddlewareHelper(...) { ... }
```

### Step 2 — Document in CHANGELOG or release notes

Every deprecation must appear in the release notes for the version that
introduces the deprecation notice.

### Step 3 — Maintain for one major version

Deprecated symbols must remain functional for at least one full major version
after the deprecation notice.

- Deprecated in v1.x → eligible for removal in v2.0.
- Deprecated in v2.x → eligible for removal in v3.0.

### Step 4 — Remove in the next major version

Removal happens only at a major version boundary. The removal must be listed in
the migration guide for that major version.

---

## Extension Packages (`x/*`) — No Compatibility Freeze

Extension packages are `experimental`. They may:

- Change API signatures between minor releases.
- Be renamed, split, merged, or removed without a major version bump.
- Graduate to stable-root status (triggering the full deprecation process from
  that point forward).

Users who depend on `x/*` packages should pin their dependency to a specific
release and review release notes on every upgrade.

When an `x/*` package graduates to a stable root, the old `x/*` path will
follow the standard deprecation process (Step 1–4 above) from that release
forward.

---

## What Is Not Deprecated by This Policy

The following changes are **not** considered breaking changes and do not require
a deprecation notice:

- Adding a new exported symbol (function, type, constant, variable).
- Adding a new optional field to a struct that uses constructor functions.
- Adding a new method to a concrete type.
- Adding a new case to a `switch` on an opaque error type.
- Fixing a bug where the old behavior was incorrect per the documented contract.
- Changing the behavior of unexported (internal) code.
- Improving error messages without changing error codes or HTTP status codes.

---

## Governance

- The package-level `module.yaml` for each stable root lists its owner and risk
  level.
- Deprecation decisions for high-risk packages (`core`, `router`, `contract`,
  `security`) require review by the runtime or security owner listed in
  the owning `module.yaml`.
- Deprecation decisions for medium-risk packages can be made by the module
  owner.
- All deprecations must pass the standard release gate before release:
  `make gates`.

---

## Current Deprecations

None at v1.0.

---

## See Also

- `docs/ROADMAP.md` — phased development plan
- `docs/stable-api/README.md` — stable-root exported API baseline and snapshot regeneration
- `docs/release/PRE_V1_RELEASE_CHECKLIST.md` — release gate evidence checklist
- `docs/EXTENSION_STABILITY_POLICY.md` — criteria for promoting `x/*` modules from experimental to beta or ga
- `<module>/module.yaml` — module owners, risk levels, and default validation
- `docs/CANONICAL_STYLE_GUIDE.md` — coding standards
- `AGENTS.md` — agent workflow and quality gates
