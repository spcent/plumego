# Safe Refactor Zones

This document defines **where refactoring is allowed**, **where it is restricted**, and **where it is effectively forbidden**.

The goal is to protect Plumego’s long-term stability while allowing safe evolution.

---

## Zone A — Free Refactor (Low Risk)

Changes are generally safe here if behavior is preserved.

### Allowed actions
- Renaming internal helpers
- Reorganizing private code
- Improving readability
- Reducing duplication
- Adding tests and examples

### Typical directories
```text
docs/
examples/
internal/
```

Public APIs must not be introduced accidentally.

---

## Zone B — Constrained Refactor (Medium Risk)

Changes are allowed but must be **minimal, explicit, and well-verified**.

### Allowed actions

* Small refactors with identical behavior
* Performance improvements with benchmarks
* Bug fixes with regression tests
* Internal restructuring without API changes

### Restrictions

* No public API changes
* No semantic behavior changes without explicit plan
* Must update tests and examples

### Typical directories

```text
router/
middleware/
context/
```

---

## Zone C — API Boundary (High Risk)

These areas define Plumego’s **public contract**.

### Rules

* Refactoring is strongly discouraged
* Any public API change requires:

  * Explicit design plan
  * Migration notes
  * Updated examples
  * Versioning consideration

### Typical directories

```text
core/
public/
```

Even internal-looking changes here can have wide impact.

---

## Zone D — Do Not Refactor (Without RFC)

Changes require an explicit RFC and maintainer approval.

### Characteristics

* Defines framework identity
* High ecosystem impact
* Breaking change potential

### Examples

* Core lifecycle semantics
* Router matching rules
* Middleware execution order guarantees

---

## Summary Table

| Zone | Risk     | Refactor Freedom | Review Strictness |
| ---- | -------- | ---------------- | ----------------- |
| A    | Low      | High             | Normal            |
| B    | Medium   | Limited          | High              |
| C    | High     | Very Limited     | Very High         |
| D    | Critical | RFC Required     | Explicit Approval |

---

## Core Rule

> The closer code is to the public surface, the less freedom an agent has to refactor it.
