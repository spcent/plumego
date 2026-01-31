# Agent Working Rules

## Plan → Implement → Verify → Rollback

This document defines the **mandatory working lifecycle** for any code agent operating on the Plumego codebase.

Plumego treats agents as **bounded engineering collaborators**, not autonomous refactoring tools.

Every non-trivial change MUST follow this sequence.

---

## 1. PLAN (Required)

No implementation may begin without an explicit plan.

### 1.1 What a valid plan includes

A plan MUST clearly state:

- **Objective**
  - What problem is being solved?
  - Is this a bug fix, feature, refactor, or documentation change?

- **Scope**
  - Which directories and packages will be touched?
  - Which files are explicitly *out of scope*?

- **Public API impact**
  - Does this change affect exported symbols?
  - Is backward compatibility preserved?

- **Risk assessment**
  - What could break?
  - Which behaviors must remain unchanged?

- **Verification strategy**
  - Which tests, benchmarks, or examples will be used to validate correctness?

### 1.2 Plan format (recommended)

```text
Objective:
Scope:
Out of scope:
Public API impact:
Risks:
Verification:
Rollback strategy:
```

If a plan cannot be written clearly, the change is likely too vague or too large.

---

## 2. IMPLEMENT (Bounded Execution)

Implementation must strictly follow the approved plan.

### 2.1 Execution rules

* Do not expand scope during implementation
* Do not refactor unrelated code “while here”
* Prefer small, reviewable diffs
* One concern per commit

### 2.2 Coding principles

* Prefer explicit code over clever abstractions
* Avoid reflection, magic defaults, or hidden behavior
* Preserve existing semantics unless explicitly planned otherwise
* Match existing style and naming conventions

---

## 3. VERIFY (Evidence-Based)

A change is not complete until it is verified.

### 3.1 Required verification

At least one of the following MUST be provided:

* Passing unit tests
* Passing golden tests
* Updated and passing examples
* Benchmarks showing no regression
* `doctor` / dump output comparison (if applicable)

### 3.2 Verification evidence

Verification results must be **visible and reviewable**, such as:

* Test output logs
* Benchmark before/after comparison
* Example output diffs
* API snapshot diffs

Assertions without evidence are insufficient.

---

## 4. ROLLBACK (Mandatory Safety Net)

Every change must be reversible.

### 4.1 Rollback requirements

The agent must explicitly state:

* How to revert the change (commit, flag, config)
* Whether rollback restores exact previous behavior
* Any data or compatibility risks during rollback

If rollback is unclear or unsafe, the change is not ready.

---

## 5. Summary Rule

> **No plan → no code**
> **No verification → no merge**
> **No rollback path → no approval**

Predictability and correctness are more important than speed.
