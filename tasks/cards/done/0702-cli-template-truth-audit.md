# Card 0702

Milestone: M-001
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: cmd/plumego
Owned Files: cmd/plumego/commands/new.go, cmd/plumego/internal/scaffold/scaffold.go, cmd/plumego/README.md, docs/getting-started.md, tasks/cards/done/0702-cli-template-truth-audit.md
Depends On: none

Goal:
Create a single source-of-truth audit for `plumego new` templates so CLI behavior, scaffold implementation, and docs can be aligned in follow-up cards.

Scope:
Compare the template names accepted by `cmd/plumego/commands/new.go` with template generation support in `cmd/plumego/internal/scaffold/scaffold.go` and the template lists in `cmd/plumego/README.md` and `docs/getting-started.md`.

Non-goals:
Do not change CLI behavior, scaffold output, or docs in this card.
Do not add new template names.

Files:
cmd/plumego/commands/new.go
cmd/plumego/internal/scaffold/scaffold.go
cmd/plumego/README.md
docs/getting-started.md
tasks/cards/done/0702-cli-template-truth-audit.md

Tests:
go test -timeout 20s ./cmd/plumego/...

Docs Sync:
Record the audit result in this card under Outcome, including supported CLI templates, scaffold templates, documented templates, and mismatches.

Done Definition:
The template truth matrix is recorded in Outcome.
The next implementation card has a clear allowlist decision.
No runtime behavior changes are made.

Outcome:
Template truth matrix:

| Template | CLI allowlist | Scaffold files | Scaffold content | CLI README | Getting started | Decision |
| --- | --- | --- | --- | --- | --- | --- |
| `canonical` | yes | yes | yes | yes | yes | keep supported |
| `minimal` | yes | yes | yes | yes | implicit default | keep supported |
| `api` | yes | yes | yes | yes | yes | keep supported |
| `fullstack` | yes | yes | yes | yes | no | keep supported |
| `microservice` | yes | yes | yes | no | no | keep supported |
| `rest-api` | no | yes | yes | yes | yes | add to CLI allowlist |
| `tenant-api` | no | yes | yes | yes | yes | add to CLI allowlist |
| `gateway` | no | yes | yes | yes | yes | add to CLI allowlist |
| `realtime` | no | yes | yes | yes | yes | add to CLI allowlist |
| `ai-service` | no | yes | yes | yes | yes | add to CLI allowlist |
| `ops-service` | no | yes | yes | yes | yes | add to CLI allowlist |

Mismatch summary:

- `cmd/plumego/commands/new.go` validates only `canonical`, `minimal`, `api`, `fullstack`, and `microservice`.
- `cmd/plumego/internal/scaffold/scaffold.go` has file lists and content dispatch for `rest-api`, `tenant-api`, `gateway`, `realtime`, `ai-service`, and `ops-service`.
- `cmd/plumego/README.md` and `docs/getting-started.md` document scenario templates that the CLI currently rejects before scaffold execution.

Allowlist decision for Card 0703:

- Add all scaffold-backed scenario templates to the CLI allowlist.
- Update the CLI help/error template list from a hand-written subset to the complete supported set.
- Preserve invalid-template fail-closed behavior with a non-zero result.

Validation:

- Not run for this analysis-only card; Card 0703 owns behavior and CLI test updates.

