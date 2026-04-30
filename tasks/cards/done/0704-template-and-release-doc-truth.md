# Card 0704

Milestone: M-001
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: docs
Owned Files: README.md, README_CN.md, docs/ROADMAP.md, docs/getting-started.md, cmd/plumego/README.md
Depends On: 0703-cmd-plumego-template-allowlist

Goal:
Update template and release-status documentation so it describes only behavior and version claims that are reproducible from the repository.

Scope:
Synchronize template lists in user-facing docs with the fixed CLI behavior, and remove or downgrade release claims that lack tag or command evidence.

Non-goals:
Do not change CLI code.
Do not add new roadmap commitments.
Do not label any `x/*` module stable or GA.

Files:
README.md
README_CN.md
docs/ROADMAP.md
docs/getting-started.md
cmd/plumego/README.md

Tests:
go run ./internal/checks/agent-workflow
go run ./internal/checks/reference-layout
cd cmd/plumego && go run . new --template rest-api --dry-run trust-check

Docs Sync:
This card is docs sync. Keep statements factual and tied to executable commands or existing repository artifacts.

Done Definition:
Documented templates match `plumego new --template` behavior.
Release status statements match repository tags and known verification facts.
No unverified RC or GA claim remains.

Outcome:
Completed.

Changes:

- Replaced unverified `v1.0.0-rc.1` badges with `pre-v1` status badges in root READMEs.
- Reworded support matrices to describe current repository status before a tagged v1 release.
- Removed stable-root `GA` wording from current-status documentation.
- Added the full supported `plumego new --template` set to the CLI README.
- Updated getting-started scaffold text to distinguish baseline templates from scenario templates.
- Added a roadmap note that release status must stay tied to tags and gate output.

Validation:

- `go run ./internal/checks/agent-workflow` passed.
- `go run ./internal/checks/reference-layout` passed.
- `cd cmd/plumego && go run . new --template rest-api --dry-run trust-check` passed with a local `GOCACHE` workaround after the default Go cache was blocked by sandbox permissions.
- `rg -n "v1.0.0|rc\\.1|\\bGA\\b|v1 release scope|v1 发布范围" ...` reports only the roadmap non-goal line forbidding `x/*` GA promotion without evidence.

