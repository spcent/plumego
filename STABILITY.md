# Stability and Compatibility Policy

This document defines compatibility guarantees for Plumego users and the maturity levels for all packages.

## Quick Reference

**Stable Roots (GA):** `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`
- **Guarantee:** Full `v1` API compatibility; breaking changes require major version bump.
- **For:** Production adoption with zero breaking changes expected within `v1.x`.

**Beta Extensions:** `x/frontend`, `x/gateway`, `x/messaging`, `x/observability`, `x/rest`, `x/tenant`, `x/websocket`
- **Guarantee:** Stable across cited release refs; breaking changes require deprecation notice first.
- **For:** Production adoption with understanding that edge cases may still be rough.

**Experimental Extensions:** All other `x/*` packages
- **Guarantee:** APIs may change in any minor version without notice.
- **For:** Pilot projects, integration tests, and explicit project-level adoption decisions.

---

## Stable Root Packages

The following nine packages carry full `v1` General Availability (GA) guarantees:

| Package | Role | Owner |
|---|---|---|
| `core` | App construction, route registration, middleware attachment, server lifecycle | core-team |
| `router` | Route matching, path parameters, groups, metadata, reverse URL generation | routing |
| `contract` | Response writers, structured error builders, request metadata, transport binding | contract |
| `middleware` | Transport-only middleware composition and first-party bundled middleware | core-team |
| `security` | Authentication, JWT, passwords, security headers, input safety, abuse guards | security |
| `store` | Storage contracts and in-memory primitives (cache, KV, file, DB, idempotency) | persistence |
| `health` | Health check and readiness models for app and dependency status | operations |
| `log` | Logging interfaces and default implementation | operations |
| `metrics` | Metrics contracts (counters, gauges, timings, collectors) | operations |

### Stability Guarantee

- **API Frozen:** Exported types, functions, interfaces, and method signatures are frozen in `v1.x`.
- **Behavior Frozen:** Package behavior matches the documented contract and does not change in breaking ways.
- **Removal:** Only through explicit migration path and deprecation cycle.
- **Breaking Changes:** Require a major version bump (e.g., `v1.x` → `v2.0`).

### You Can Safely

- Import and update to any `v1.x` release.
- Build services with no concern for API churn.
- Expect stable test suites that validate compatibility across releases.

---

## Extension Maturity Levels

All `x/*` packages are classified by maturity. Maturity determines API stability guarantees and production readiness.

### Experimental

**Definition:** API design is still evolving; backwards compatibility is **not** guaranteed.

**When to use:**
- Integration and pilot projects.
- Explicit, project-level adoption decisions.
- When the feature is mission-critical and the team owns the migration risk.

**When not to use:**
- Production services expecting zero API churn.
- Shared dependencies across multiple teams or services.
- Long-term code with minimal expected maintenance.

**API Expectations:**
- Exported symbols may be added, removed, or renamed in any minor version.
- Method signatures may change.
- Behavior may change without deprecation notice.
- No breaking-change policy applies.

**Current experimental packages:** `x/ai`, `x/data`, `x/fileapi`, `x/openapi`, `x/resilience`, `x/rpc`, `x/validate`, plus experimental surfaces within `x/gateway`, `x/messaging`, and `x/observability`.

**Promotion Path:** Experimental → Beta → GA (see [Promotion Criteria](#promotion-criteria)).

---

### Beta

**Definition:** API is stable across cited release references; suitable for production adoption with minor caveats acknowledged.

**When to use:**
- Production services with acceptable edge-case risk.
- Teams ready to manage minor breaking changes with deprecation notice.
- Features essential to your product where alternatives don't exist.

**When not to use:**
- Ultra-stable systems where any breaking change is disruptive (use stable roots instead).
- Short-lived projects where churn is unacceptable (use stable roots instead).

**API Expectations:**
- Exported API is stable across the cited release references in the beta evidence document.
- Breaking changes require a deprecation notice in the PR that introduces the change.
- Deprecated symbols must be removable in a subsequent minor version.
- Backwards compatibility is the default expectation, with clear migration paths when changes occur.

**Current beta packages:** `x/frontend`, `x/gateway`, `x/messaging`, `x/observability`, `x/rest`, `x/tenant`, `x/websocket`.

**Evidence:** Each beta package has a corresponding evidence document in `docs/evidence/extension/` demonstrating API stability across release references.

**Promotion Path:** Beta → GA (see [Promotion Criteria](#promotion-criteria)).

---

### General Availability (GA)

**Definition:** API carries the same guarantee as stable root packages; production-ready with no breaking changes expected within a major version.

**When to use:**
- Any production system requiring zero API churn.
- Shared libraries and frameworks.
- Long-lived applications.

**API Expectations:**
- Same as stable root packages.
- Breaking changes require a major version bump.
- Backwards compatibility is guaranteed within `v1.x`.

**Current GA packages:** Stable root packages only (see [Stable Root Packages](#stable-root-packages)).

**Promotion Path:** Packages are promoted to GA after two production release cycles with zero exported API changes and full test coverage.

---

## SemVer Expectations

Plumego follows Semantic Versioning for version numbers: `MAJOR.MINOR.PATCH`.

| Surface | v1.x Guarantee | Breaking-Change Path |
|---|---|---|
| **Stable roots** | No breaking changes within `v1` | Major version bump (`v2+`) |
| **Beta extensions** | Stable across cited release refs in beta evidence | Deprecation notice in change PR → removal in subsequent minor |
| **Experimental extensions** | No guarantee; API may change in any minor version | None required; changes permitted freely |
| **CLI (`cmd/plumego`)** | Command-line interface stability; not a Go import surface | Minor version change with notice |

### Version Numbers

- **MAJOR:** Incremented on breaking changes to stable roots or GA extensions.
- **MINOR:** Incremented on new features, new experimental packages, or beta promotion events.
- **PATCH:** Incremented on bug fixes and internal improvements.

---

## Go Version Support Policy

Plumego targets recent Go versions and maintains compatibility for **two most recent stable Go releases** in `v1.x`.

| Plumego Version | Minimum Go Version | Tested Against |
|---|---|---|
| v1.1.0 | Go 1.24 | Go 1.24, 1.25, 1.26 (beta at release) |
| v1.2+ | Go 1.25 | Go 1.25, 1.26+ |

**Rationale:** Plumego's design leverages modern stdlib features; supporting multiple Go versions older than two releases introduces maintenance burden with diminishing benefit.

---

## Breaking Change Policy

A breaking change is any modification to stable roots or GA extensions that requires code changes in consuming packages.

### Examples of Breaking Changes

- Removing an exported function, type, interface, or method.
- Changing a function signature (parameter order, types, or count).
- Changing a struct field name or type.
- Changing an interface method signature.
- Removing an interface method.
- Changing the semantics or behavior of documented functionality.
- Changing error types or error return values.

### Examples of Non-Breaking Changes

- Adding a new exported function, type, or method.
- Adding a new struct field (if it does not require changes to callers).
- Adding a new method to an exported interface (if interface implementations are not part of the public contract).
- Improving performance.
- Fixing a bug (unless the bug was depended upon).

### Migration and Removal Strategy

For stable roots and GA extensions:

1. **Announce deprecation:** Mark the symbol with a deprecation comment and update release notes.
2. **Provide migration path:** Document the replacement and any rewrite tool or guide.
3. **Support window:** Maintain for at least two minor versions (e.g., v1.x and v1.(x+1)).
4. **Remove:** Only after full migration of in-repo callers; no dead wrappers at merge time.

---

## Deprecation Policy

Deprecated symbols are marked with a deprecation comment block and recorded in `docs/evidence/deprecation-inventory.md`.

### Deprecation Marker

```go
// Deprecated: Use NewFoo instead. Remove in v2.0.
func OldFoo() { }
```

### Removal Requirement

- Deprecated symbols within stable roots must be removed only when a major version bump occurs.
- Deprecated symbols within beta extensions must be removed in a subsequent minor version after the deprecation PR.
- No dead wrappers or migration helpers may remain at merge time; all in-repo callers must be migrated in the same PR.

---

## What You Can Safely Depend On

### Safe for Production

- **Stable root packages:** Import any stable root and update to any `v1.x` release without concern.
- **Beta extensions:** Adopt beta extensions understanding that breaking changes are rare but may occur with deprecation notice.
- **Documented behavior:** Any behavior documented in `README.md`, module primers, or formal docs.

### Do Not Treat as Stable

- **Experimental `x/*` extensions:** APIs may change in any minor version without notice.
- **Internal packages (`internal/`):** Private implementation detail; not subject to stability guarantees.
- **`cmd/plumego` as a Go import surface:** It is a CLI tool; not an importable library.
- **Undocumented behavior:** Behavior not covered by docs or examples has no stability guarantee.

---

## Promotion Criteria

### Experimental → Beta

Requirements for a package to move from experimental to beta:

1. **API stability proof:** No exported API changes across two consecutive release references.
2. **Evidence documentation:** Complete entry in `docs/evidence/extension/` with owner sign-off.
3. **Module manifest update:** Add status field to the extension's `module.yaml`.
4. **Test coverage:** Minimum 70% coverage at stable tier (`go test -cover`).
5. **Documentation:** Complete module primer in `docs/modules/x/<package>/README.md`.

### Beta → GA

Requirements for a package to move from beta to GA:

1. **Release cycles:** Two full production release cycles with zero exported API changes.
2. **Coverage:** Full coverage at stable tier.
3. **Boundary review:** Sign-off that module respects core-boundary contracts.
4. **Stable roots treatment:** Promote to `stable` root package (move from `x/*` to top-level).
5. **Evidence:** Updated `docs/evidence/stable-api/` entry.

---

## Contact and Updates

For questions about stability, maturity, or promotion:

1. Check the [Extension Maturity Dashboard](./docs/concepts/extension-maturity.md).
2. Review the beta evidence for your package in `docs/evidence/extension/`.
3. Consult the module primer in `docs/modules/`.
4. Open an issue on [GitHub](https://github.com/spcent/plumego).

---

## Policy Version

This policy is authoritative as of Plumego v1.1.0 and applies to all v1.x releases.

Changes to this policy require a minor version release and explicit documentation of the change in release notes.
