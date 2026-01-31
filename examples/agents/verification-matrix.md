# Verification Matrix

This document defines **mandatory verification steps** based on change type.

Agents must use this matrix to determine what evidence is required before approval.

---

## Change Types

- **Bugfix**
- **Feature**
- **Refactor**
- **Breaking Change**
- **Documentation**

---

## Verification Requirements

### Bugfix

Required:
- Unit test reproducing the bug
- Test passing after fix

Recommended:
- Regression test
- Example update (if user-facing)

---

### Feature

Required:
- Unit tests or examples
- Documentation update
- No regression in existing tests

Recommended:
- Benchmark (if performance-sensitive)
- Golden test (for routing/middleware)

---

### Refactor (No Behavior Change)

Required:
- All existing tests passing
- Example outputs unchanged
- Diff reviewable and scoped

Recommended:
- Golden test comparison
- Before/after behavior snapshot

---

### Breaking Change

Required:
- Explicit plan and approval
- Updated examples
- Migration guide
- API snapshot diff
- Versioning justification

Recommended:
- Compatibility layer (temporary)
- Deprecation notice

---

### Documentation Change

Required:
- Content accuracy check
- Internal link validation

Recommended:
- Example verification (if referenced)

---

## Verification Tools (Examples)

- `go test ./...`
- Golden test diff
- Benchmark comparison
- Example output verification
- API symbol snapshot diff

---

## Non-Negotiable Rule

> If verification is unclear, the change is incomplete.