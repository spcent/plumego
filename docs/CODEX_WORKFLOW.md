# Codex Workflow — Plumego Autonomous Evolution

This document describes how plumego uses `codex --yolo` as its execution engine
so that human effort concentrates entirely on writing Milestone Specs and
reviewing Milestone PRs.

## Mental Model

```
┌─────────────────────────────────────────────────────────┐
│  Human Work Surface                                     │
│  ─────────────────                                      │
│  1. Write Milestone Spec  →  tasks/milestones/active/   │
│  2. Review Milestone PR   ←  GitHub PR                  │
└──────────────┬──────────────────────────────────────────┘
               │  make milestone M=active/M-NNN
               ▼
┌─────────────────────────────────────────────────────────┐
│  Codex --yolo (autonomous execution)                    │
│  ──────────────────────────────────                     │
│  reads AGENTS.md  →  reads spec context  →  executes   │
│  tasks  →  runs quality gates  →  commits  →  pushes   │
└──────────────┬──────────────────────────────────────────┘
               │  git push milestone/M-NNN-slug
               ▼
┌─────────────────────────────────────────────────────────┐
│  CI (GitHub Actions: milestone-gates.yml)               │
│  ─────────────────────────────────────────              │
│  dependency-rules  agent-workflow  module-manifests     │
│  reference-layout  go vet  gofmt  tests  tests-race     │
│  → posts gate table as PR comment                       │
└─────────────────────────────────────────────────────────┘
```

The human never touches a terminal during a milestone run.  
The only manual node is the PR review.

---

## Roles and Responsibilities

| Role   | Responsibility                                               |
|--------|--------------------------------------------------------------|
| Human  | Write specs. Review PRs. Merge. Archive spec to `done/`.    |
| Codex  | Execute spec tasks. Run gates. Commit. Push. Open PR.        |
| CI     | Re-run gates. Post results. Block merge on failure.         |

---

## One-Time Local Setup

Install the pre-push git hook so quality gates run locally before every push
(including when Codex pushes a milestone branch):

```bash
make setup-hooks
```

After this, `git push` on a `milestone/*` branch automatically runs the full
gate suite and blocks the push if any gate fails — identical to CI, but instant
and without waiting for GitHub Actions.

| Branch pattern | Local gate triggered |
|----------------|----------------------|
| `milestone/*`  | Full suite (all 8 gates) |
| `main`, `claude/*`, `feature/*` | Quick check (vet + fmt + tests) |
| anything else  | No gate (push proceeds) |

To skip the hook for a one-off push: `git push --no-verify`

---

## Step-by-Step Workflow

### 1. Write a Milestone Spec

Copy `tasks/milestones/TEMPLATE.md` to `tasks/milestones/active/M-NNN.md`.

Fill in:

- **Goal** — one sentence.
- **Context** — files Codex reads before touching code. Always include the target
  module's `module.yaml` and the relevant `specs/` files.
- **Affected Modules** — primary module, at most one secondary.
- **Tasks** — ordered, atomic steps. Keep each step small enough that it
  produces a diff you can review in isolation.
- **Acceptance Criteria** — the exact shell commands that must exit 0.
  Do not soften or abbreviate them.
- **Out of Scope** — explicit stops.

Commit the spec file to `main` (or a feature branch). This is the only code
humans write for this milestone.

### 2. Launch Codex

```bash
make milestone M=active/M-NNN
```

This runs:

```bash
codex --yolo "$(cat tasks/milestones/active/M-NNN.md)"
```

Codex reads `AGENTS.md` automatically (it is the behavior boundary for every
Codex session in this repo). The spec provides the task.

Walk away. Codex runs until the spec is done or it hits a blocker.

### 3. Watch CI

When Codex pushes the milestone branch, GitHub Actions runs
`milestone-gates.yml` (source: `docs/github-workflows/milestone-gates.yml`).
It posts a gate table to the PR.

> **First-time CI setup:** `.github/` is in `.gitignore`. Copy the workflow once:
> ```bash
> mkdir -p .github/workflows
> cp docs/github-workflows/milestone-gates.yml .github/workflows/
> git add -f .github/workflows/milestone-gates.yml
> git commit -m "ci: add milestone quality gates workflow"
> ```

If CI is red, Codex should have already fixed failures before pushing —
a red push means something was missed. You can re-run Codex on the same branch:

```bash
git checkout milestone/M-NNN-slug
codex --yolo "Fix all failing quality gates. Gates are listed in AGENTS.md section 8."
```

### 4. Review the PR

The PR is the human checkpoint. Review:

1. CI gate table is all green.
2. Diff is within the spec's **Affected Modules** and **Out of Scope** list.
3. No stable root imports `x/*` (enforced by `dependency-rules` check).
4. No new entries in `go.mod` (main module must stay stdlib-only).
5. Tests cover the changed behavior.

Merge when satisfied.

### 5. Archive the Spec

Move the spec from `active/` to `done/`:

```bash
mv tasks/milestones/active/M-NNN.md tasks/milestones/done/M-NNN.md
```

Add an `## Outcome` section:

```markdown
## Outcome

- PR: #NNN
- Merged: YYYY-MM-DD
- Gates: all pass (see PR comment)
```

Commit directly to `main`.

---

## Writing High-Signal Specs

### Context is the Highest-Leverage Field

Codex performs better when you list the exact files to read, not just module names.
Prefer:

```markdown
3. `x/rest/module.yaml`
4. `specs/change-recipes/add-http-endpoint.yaml`
5. `reference/standard-service/internal/app/routes.go`
```

over:

```markdown
3. x/rest docs
```

### Tasks Must Be Atomic

Each task step should change at most one file or produce one testable unit.
If you find yourself writing "and then also", split it into two steps.

### Acceptance Criteria Must Be Exact

Never write "tests should pass". Write the exact command:

```bash
go test -race -timeout 60s ./x/rest/...
```

Codex will run exactly what you write. Vague criteria produce vague validation.

### Out of Scope Prevents Scope Creep

Codex in `--yolo` mode is eager. Explicit stops prevent it from tidying unrelated
code, refactoring adjacent packages, or adding features not in the spec.

Common stops to include:
- "Do not change stable root public APIs."
- "Do not add entries to go.mod."
- "Do not modify `reference/standard-service` unless this spec explicitly lists it."

---

## Parallel Milestones

You can queue multiple specs in `active/` simultaneously, but run them sequentially
if they touch overlapping modules. The milestone branch names (`milestone/M-NNN-slug`)
prevent conflicts automatically as long as specs stay within their declared
**Affected Modules**.

---

## Spec Authoring Checklist

Before dropping a spec into `active/`:

- [ ] Goal is one sentence.
- [ ] Context lists specific file paths, not vague descriptions.
- [ ] Affected modules declared (primary + any secondaries).
- [ ] Tasks are ordered and each is atomic.
- [ ] Acceptance criteria are exact shell commands.
- [ ] Out of Scope has at least two stops.
- [ ] Spec is under ~80 lines total. If longer, split it.

---

## Behavioral Boundaries for Codex

`AGENTS.md` is the behavioral contract. It is loaded automatically by Codex on
every session. The spec does not need to re-state AGENTS.md rules. Instead, rely
on them:

- All changes respect stable root vs `x/*` boundaries (AGENTS.md §6).
- Commits follow the convention in AGENTS.md §11.
- Validation follows the gate order in AGENTS.md §8.
- Docs sync follows AGENTS.md §9.

If you want Codex to deviate from a default (e.g., touch a stable root API),
state it explicitly in the spec's **Tasks** section.

---

## Troubleshooting

| Symptom | Likely Cause | Action |
|---------|--------------|--------|
| Codex stops mid-way | Blocker found (see spec `## Blocker`) | Fix the blocker, re-run |
| CI gate `dependency-rules` fails | Stable root now imports `x/*` | Check new import paths |
| CI gate `module-manifests` fails | New file not listed in `doc_paths` | Update `module.yaml` |
| PR diff outside affected modules | Spec's **Out of Scope** was too vague | Revert, tighten spec, re-run |
| `go.mod` has new dependency | Extension added non-stdlib import to main module | Move to `x/*` sub-module |
