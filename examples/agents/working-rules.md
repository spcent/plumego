# Examples of Good Pull Requests

This document defines what a **high-quality PR** looks like in the Plumego project.

The goal is consistency, auditability, and long-term maintainability.

---

## 1. Good Commit Messages

### Characteristics
- Clear intent
- Scope-aware
- No emotional language

### Examples

Good:
```text
router: fix incorrect param decoding for encoded paths
middleware: clarify recovery behavior on panic
docs: add rationale for explicit middleware ordering
```

Bad:

```text
fix stuff
cleanup
improve router
```

---

## 2. Good Diff Size

Guidelines (not hard limits):

* **Ideal**: < 300 lines
* **Acceptable**: 300–800 lines (with justification)
* **>800 lines**: should be split or formally reviewed

Large diffs must include:

* Clear structure
* Commit separation
* Explicit rationale

---

## 3. Good Verification Evidence

A good PR includes **proof**, not claims.

Examples:

* Test output snippets
* Benchmark before/after numbers
* Golden test diffs
* Example output confirmation

Avoid:

* “Should work”
* “Looks correct”
* “No behavior change” without evidence

---

## 4. Good PR Description Template

```text
Summary:
What problem does this PR solve?

Scope:
Which areas keeps unchanged?
Which areas are modified?

Public API impact:
None / Described below

Verification:
- Tests:
- Examples:
- Benchmarks:

Rollback:
How to revert safely
```

---

## 5. Good PR Mindset

A good PR optimizes for:

* Reviewability
* Predictability
* Reversibility

Not for:

* Cleverness
* Brevity at the cost of clarity
* “While I’m here” refactors

---

## Final Principle

> A PR should make the system easier to understand than before.

If it does not, it is not done.