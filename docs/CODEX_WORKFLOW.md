# Codex Workflow — Milestone-Driven Agent Delegation

Companion lightweight pipeline spec:
[`docs/MILESTONE_PIPELINE.md`](./MILESTONE_PIPELINE.md)

## The Model

```
┌──────────────────────────────────────────────────────────────┐
│  You (Architect)                                             │
│  ─────────────────────────────────────────────────────────  │
│  Decide: goal, scope, architectural constraints,            │
│          task phases, acceptance criteria, hard stops.      │
│  Tool: make new-milestone N=NNN TITLE="..."                 │
│        → fill TEMPLATE → make check-spec → commit          │
└────────────────────────┬─────────────────────────────────────┘
                         │  make milestone M=active/M-NNN
                         ▼
┌──────────────────────────────────────────────────────────────┐
│  Codex --yolo (Autonomous Execution)                        │
│  ─────────────────────────────────────────────────────────  │
│  Reads AGENTS.md (behavior boundary, auto-loaded)           │
│  Reads spec Context files                                   │
│  Executes phases:                                           │
│    Phase 1 (sequential) → Phase 2 (parallel) → ...         │
│  Runs quality gates → commits → pushes → opens PR          │
└────────────────────────┬─────────────────────────────────────┘
                         │  git push milestone/M-NNN-slug
                         ▼
┌──────────────────────────────────────────────────────────────┐
│  Local pre-push hook (scripts/pre-push)                     │
│  Full gate suite runs before push leaves the machine        │
│  Blocks push on any failure                                 │
└────────────────────────┬─────────────────────────────────────┘
                         │  PR opened
                         ▼
┌──────────────────────────────────────────────────────────────┐
│  CI (docs/github-workflows/milestone-gates.yml)             │
│  Reruns all gates, posts result table to PR comment         │
└────────────────────────┬─────────────────────────────────────┘
                         │  CI green
                         ▼
┌──────────────────────────────────────────────────────────────┐
│  You (Reviewer) — the only manual checkpoint                │
│  ─────────────────────────────────────────────────────────  │
│  1. Goal matches intent?                                    │
│  2. Approach section confirms method?                       │
│  3. Architecture Decisions table: all ✅?                   │
│  4. Diff within declared scope?                             │
│  5. go.mod unchanged?                                       │
│  → Merge + archive spec + next milestone                    │
└──────────────────────────────────────────────────────────────┘
```

---

## One-Time Setup

```bash
# Install local pre-push quality gate hook
make setup-hooks
```

After this, quality gates run automatically on every `git push`:

| Branch | Gate triggered |
|--------|---------------|
| `milestone/*` | Full 8-gate suite (blocks push on failure) |
| `main`, `claude/*`, `feature/*` | Quick check (vet + fmt + tests) |

To skip once: `git push --no-verify`

---

## Complete Workflow

### Step 1 — Scaffold the Spec

```bash
make new-milestone N=001 TITLE="Add ResourceHandler to x/rest"
# → creates tasks/milestones/active/M-001.md from TEMPLATE
```

Fill the spec. Concentrate effort on:

| Section | What to write | Time investment |
|---------|--------------|----------------|
| **Goal** | One sentence — observable end state | Medium |
| **Architecture Decisions** | Non-obvious constraints Codex must not override | High |
| **Context** | Exact file paths, not module names | Medium |
| **Tasks / Phases** | Atomic steps; mark parallel groups explicitly | High |
| **Acceptance Criteria** | Exact commands, verbatim | Low (copy template) |
| **Out of Scope** | Hard stops; be explicit | Medium |

### Step 2 — Validate the Spec

```bash
make check-spec M=active/M-001
```

Checks all required sections and gate commands are present.  
Fix any `MISS` lines. Warnings about placeholders mean the spec is incomplete.

If you are using the full pipeline artifacts, scaffold and validate them too:

```bash
make new-plan M=active/M-001
make check-plan M=active/M-001
make new-card ID=001 SLUG=slice-router-work M=M-001
make check-card C=active/C-001-slice-router-work
make new-verify M=active/M-001
make check-verify M=active/M-001
```

Update `tasks/milestones/ROADMAP.md`: add a row for M-001, declare dependencies.

### Step 3 — Launch Codex

```bash
make milestone M=active/M-001
```

Codex reads AGENTS.md (auto-loaded) + the spec, then executes autonomously.  
Walk away. Expected timeline depends on milestone size; check back at PR.

If Codex surfaces a **Draft PR** with `## Open Questions` filled in, it hit a
decision point that needs your input. Answer in the spec, re-run.

### Step 4 — Review the PR

The PR body is filled from `docs/github-workflows/milestone-pr-template.md`.
Review in order:

1. **CI gate table** (posted by bot) — must be all green.
2. **Goal** — does it match what you intended?
3. **Approach** — is the method what you had in mind?
4. **Architecture Decisions table** — any ⚠️ rows require your explicit approval.
5. **Scope boundary table** — no unexpected modules in the diff.
6. **`go.mod`** — must be unchanged.

Merge when satisfied.

### Step 5 — Archive

```bash
mv tasks/milestones/active/M-001.md tasks/milestones/done/M-001.md
# Add ## Outcome section with PR number and date
```

Update `tasks/milestones/ROADMAP.md`: set status to `[✓]`.  
Any milestone that listed `Depends on: M-001` is now unblocked.

---

## Writing High-Signal Specs

### Architecture Decisions — the Highest-Value Section

This section is what separates a good spec from an ambiguous one.  
Without it, Codex makes reasonable but possibly wrong choices.  
With it, Codex implements exactly what you intended.

Write decisions that are:
- **Non-obvious** — Codex would not infer them from context alone.
- **Specific** — "use interface embedding, not wrapping" beats "keep http.Handler compatible".
- **Falsifiable** — a reviewer can check whether the decision was followed.

Do NOT write:
- "Follow best practices" — not actionable.
- "Keep it simple" — not falsifiable.
- Decisions that are already enforced by AGENTS.md or `specs/` — redundant.

### Context — Exact Paths Beat Module Names

```markdown
# Good
4. `x/rest/module.yaml`
5. `specs/change-recipes/add-http-endpoint.yaml`
6. `reference/standard-service/internal/app/routes.go`

# Bad
4. x/rest docs and specs
```

### Parallel Task Phases — When to Use Them

Mark a phase as parallel when its tasks:
- Touch different files with no shared state.
- Could be implemented in any order without conflicts.
- Would produce independent, reviewable diffs.

Do NOT mark as parallel when:
- Task B imports or extends what Task A creates.
- Both tasks modify the same file.

### Out of Scope — Be Explicit

Add milestone-specific stops beyond the standard ones.  
If you know a tempting-but-wrong path exists, name it:

```markdown
- Do not extract a generic `HandlerRegistry`; ResourceHandler is the only abstraction needed.
- Do not modify `reference/with-gateway`; only `reference/standard-service` is in scope.
```

---

## Delegation Contract Summary

| Concern | Who Decides | Where |
|---------|-------------|-------|
| Goal, scope, hard stops | You | Spec |
| Architectural constraints | You | `## Architecture Decisions` |
| Internal implementation details | Codex | Autonomously |
| Type/func naming (non-public) | Codex | Per existing patterns |
| Test structure | Codex | Per CANONICAL_STYLE_GUIDE |
| Commit granularity | Codex | Per AGENTS.md §11 |
| Quality gate failures | Codex | Must fix before push |
| Deviations from Architecture Decisions | Codex surfaces, You approve | PR `⚠️` row |
| New external dependencies | Neither | Forbidden |
| Touching out-of-scope modules | Neither | Forbidden |

Full contract: `AGENTS.md §12`.

---

## Parallel Milestone Execution

Multiple milestones can run concurrently when:
- Their `Parallel OK:` field is `yes`.
- Their `Depends on:` fields are both satisfied.
- Their `## Affected Modules` do not overlap.

Check `tasks/milestones/ROADMAP.md` before launching a second milestone.  
If two milestones accidentally touch the same file, run them sequentially.

---

## Troubleshooting

| Symptom | Cause | Action |
|---------|-------|--------|
| `make check-spec` shows MISS | Required section missing | Fill section, re-run |
| `make check-plan` fails | Plan missing or malformed | Run `make new-plan`, fill fields, re-run |
| `make check-card` warns about legacy shape | Card predates milestone pipeline metadata | Add milestone ownership fields or keep as legacy queue item |
| `make check-verify` fails | Verify report missing required gate sections | Scaffold with `make new-verify`, then fill evidence |
| Codex opens Draft PR with Open Questions | Hit unresolvable decision point | Answer in spec, re-run Codex on branch |
| Pre-push hook blocks push | Gate failure | Fix failure, `git push` again |
| CI gate `dependency-rules` fails | Stable root imports `x/*` | Check new import paths in diff |
| CI gate `module-manifests` fails | New file not in `doc_paths` | Update `<module>/module.yaml` |
| PR diff outside Affected Modules | Out of Scope too vague | Revert, tighten spec, re-run |
| `go.mod` has new entry | Extension added non-stdlib import to main module | Move to `x/*` sub-module |
| Two parallel milestones conflict | Overlapping files | Serialize them in ROADMAP.md |

---

## Reference

| File | Purpose |
|------|---------|
| `tasks/milestones/TEMPLATE.md` | Spec template |
| `tasks/milestones/PLAN_TEMPLATE.md` | Planner handoff template |
| `tasks/milestones/VERIFY_TEMPLATE.md` | Verifier evidence template |
| `tasks/milestones/ROADMAP.md` | Pipeline view, dependency DAG |
| `tasks/cards/TEMPLATE.md` | Worker card template |
| `docs/github-workflows/milestone-pr-template.md` | PR body template (Codex fills) |
| `docs/github-workflows/milestone-gates.yml` | CI workflow (copy to `.github/workflows/`) |
| `scripts/check-spec` | Spec validator |
| `scripts/pre-push` | Local pre-push gate hook |
| `AGENTS.md §11` | Milestone execution protocol |
| `AGENTS.md §12` | Delegation contract |
| `Makefile` | `make help` for all targets |
