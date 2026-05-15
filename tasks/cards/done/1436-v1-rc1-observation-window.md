# Card 1436

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: release
Owned Files:
- `docs/release/v1.0.0.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/M-005.verify.md`
- `tasks/cards/active/README.md`
Depends On:
- 1434

Goal:
- Complete the `v1.0.0-rc.1` observation window and decide whether final
  `v1.0.0` can be tagged or whether an `rc.2` blocker set is required.

Scope:
- Re-check GitHub Actions status for `v1.0.0-rc.1`.
- Review open P0/P1 issues, release blockers, and any regressions reported
  during the observation window.
- If clean, update final release notes for GO.
- If blocked, create one bounded card per blocker and keep final v1 untagged.

Non-goals:
- Do not tag final `v1.0.0` without a clean observation result.
- Do not promote experimental extensions.
- Do not change runtime behavior from the observation card.

Files:
- `docs/release/v1.0.0.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/M-005.verify.md`
- `tasks/cards/active/README.md`

Tests:
- `gh run view 25920615874 --repo spcent/plumego --json status,conclusion,url`
- `go run ./internal/checks/extension-beta-evidence`
- `git status --short --branch`

Docs Sync:
- Required for final GO/NO-GO status and any blocker cards.

Done Definition:
- Observation result is recorded as GO or NO-GO.
- Final `v1.0.0` remains untagged unless GO is recorded.
- Any blocker has a bounded active card.

Outcome:
- Remote rc gate remained green:
  - `gh run view 25920615874 --repo spcent/plumego --json status,conclusion,url,headSha,event,displayTitle`
  - conclusion: success
  - head SHA: `a234681372e9e6c08abdce10048cb37be4a0a5a5`
  - URL: `https://github.com/spcent/plumego/actions/runs/25920615874`
- Open issue review found no reported blockers:
  - `gh issue list --repo spcent/plumego --state open --limit 100 --json number,title,labels,updatedAt,url`
  - result: `[]`
- Open PR review found no release-blocker PR:
  - `gh pr list --repo spcent/plumego --state open --limit 50 --json number,title,labels,headRefName,updatedAt,url`
  - result: PR `#241 Add framework comparison and stability pages` on branch
    `claude/redesign-plumego-website-8AEbL`, with no release-blocker label.
- Extension beta evidence remained locked:
  - `go run ./internal/checks/extension-beta-evidence`
  - result: PASS
- Final local release gate passed after card 1437:
  - `GOCACHE=/private/tmp/plumego-gocache GOMODCACHE=/private/tmp/plumego-gomodcache make gates`
  - result: PASS
  - stable-module coverage: 87.2%
- Final observation result: GO.
- Final `v1.0.0` can be tagged from the committed release decision state.
