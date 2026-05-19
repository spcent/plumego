# Card 1438

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: release
Owned Files:
- `docs/release/v1.0.0.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`
- `tasks/cards/done/1438-v1-final-tag-remote-evidence.md`
Depends On:
- 1436
- 1437

Goal:
- Record final `v1.0.0` tag and remote GitHub Actions evidence.

Scope:
- Verify the final tag object, target, and remote Actions result.
- Record the current `main` run result after the release-control-plane content
  landed there.

Non-goals:
- Do not rewrite the published tag.
- Do not promote experimental extensions.
- Do not change runtime behavior.

Files:
- `docs/release/v1.0.0.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`

Tests:
- `gh run view 25922384589 --repo spcent/plumego --json status,conclusion,url,headSha,event,displayTitle`
- `gh run view 25922373632 --repo spcent/plumego --json status,conclusion,url,headSha,event,displayTitle`
- `git ls-remote --tags origin 'v1.0.0' 'v1.0.0^{}'`
- `git status --short --branch`

Docs Sync:
- Required for final release evidence.

Done Definition:
- Final tag exists remotely.
- Final tag GitHub Actions run is traceable and successful.
- Main branch evidence run is traceable and successful.

Outcome:
- Remote final tag exists:
  - tag object: `cde6de1d34c836584c54ba0df86dab3cf92ba6e2`
  - tag target: `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96`
- Final tag GitHub Actions run passed:
  - run: `25922384589`
  - URL: `https://github.com/spcent/plumego/actions/runs/25922384589`
- Current `main` evidence run passed:
  - commit: `f684bc8754d66fdfd69da07b4378e66787561395`
  - run: `25922373632`
  - URL: `https://github.com/spcent/plumego/actions/runs/25922373632`
- The first push also attempted to update `codex/v1-release-execution`; GitHub
  rejected that branch update as non-fast-forward, while the final tag push
  succeeded. No force-push was performed.
