<!--
File: .github/pull_request_template.md

This template enforces Plumego agent rules:
- Plan → Implement → Verify → Rollback
- Safe refactor zones
- Verification matrix
- Evidence-based review

Keep the PR description concise but complete.
-->

# Summary
<!-- What problem does this PR solve? What user-visible or internal behavior changes? -->

- 

## Type of change
<!-- Select exactly one primary type. -->
- [ ] Bugfix
- [ ] Feature
- [ ] Refactor (no behavior change)
- [ ] Breaking change
- [ ] Documentation

## Scope and boundaries
<!-- List the exact areas you changed. Be explicit about what you did NOT touch. -->

**Touched (directories/packages):**
- 

**Out of scope (explicitly not touched):**
- 

## Safe refactor zone declaration
<!-- Mark the highest-risk zone you touched. -->
- [ ] Zone A — Free Refactor (docs/, examples/, internal/)
- [ ] Zone B — Constrained Refactor (router/, middleware/, context/)
- [ ] Zone C — API Boundary (core/, public surface)
- [ ] Zone D — Do Not Refactor (RFC required)

**If Zone C or D:**  
- [ ] I have included a design note / RFC link and have maintainer approval.

## Plan
<!-- Required for any non-trivial change. If the change is trivial, say why. -->

**Objective:**
- 

**Approach (high level):**
- 

**Public API impact:**
- [ ] None
- [ ] Yes (describe below)

If public API changed, list exported symbols added/removed/modified:
- 

**Risks / invariants (what must NOT change):**
- 

## Verification (evidence required)
<!-- Use the Verification Matrix. Paste real outputs or link to CI runs. -->

### Required checks (select all that apply and provide evidence)
- [ ] `go test ./...` (attach output / CI link)
- [ ] Examples updated and verified (attach output / CI link)
- [ ] Golden tests updated and verified (attach diff / CI link)
- [ ] Benchmarks run (attach before/after numbers)
- [ ] API snapshot checked (attach diff) *(required for breaking/public API changes)*
- [ ] Link validation / docs checks (for docs-only PRs)

**Evidence (paste logs or link):**
```text
PASTE OUTPUT OR CI LINKS HERE
```