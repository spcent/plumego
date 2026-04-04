# M-XXX: <Title>

<!-- Replace M-XXX with the milestone number, e.g. M-007 -->
<!-- Branch: milestone/M-XXX-short-slug -->

## Goal

<!-- One sentence. What must be true when this milestone is complete? -->
<!-- Example: "x/rest exposes a canonical ResourceHandler interface that -->
<!-- replaces the three ad-hoc CRUD patterns currently in reference apps." -->

## Context — Read Before Touching Code

<!-- List the files Codex must read first, in order. -->
<!-- Always include AGENTS.md (Codex loads this automatically). -->
<!-- Add the module manifest and any relevant spec or recipe file. -->

1. `AGENTS.md` (loaded automatically)
2. `specs/dependency-rules.yaml`
3. `specs/agent-entrypoints.yaml`
4. `<primary-module>/module.yaml`
5. <!-- add more as needed -->

## Affected Modules

<!-- Primary module (one): -->
- `<primary-module>`

<!-- Secondary modules (only if unavoidable): -->
<!-- - `<secondary-module>` -->

## Tasks

<!-- Ordered. Each step must be atomic and independently testable. -->
<!-- Use imperative verb phrases. -->

1. Read all context files listed above.
2. <!-- e.g. "Add ResourceHandler interface to x/rest/handler.go" -->
3. <!-- e.g. "Update x/rest/module.yaml doc_paths to include new file" -->
4. <!-- e.g. "Add unit tests in x/rest/handler_test.go" -->
5. Run the acceptance criteria commands below and fix any failures.
6. Commit all changes with message: `feat(<module>): <short description> [M-XXX]`
7. Push to branch `milestone/M-XXX-<slug>`.

## Acceptance Criteria

<!-- These are the exact commands that must exit 0. Do not abbreviate. -->

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./<primary-module>/...
go test -timeout 20s ./...
go vet ./...
gofmt -l .
```

Expected: all commands exit 0, `gofmt -l` outputs nothing.

## Out of Scope

<!-- Hard stops. List anything that might seem tempting but must not happen. -->

- Do not change stable root public APIs.
- Do not add dependencies to the main module.
- Do not refactor unrelated packages.
- <!-- add milestone-specific stops -->

## Done Definition

<!-- The milestone is done when: -->
- [ ] All acceptance criteria commands pass with exit 0.
- [ ] `gofmt -l .` produces no output.
- [ ] The branch `milestone/M-XXX-<slug>` is pushed.
- [ ] A PR is open targeting `main` with title `milestone(M-XXX): <Title>`.
- [ ] PR description contains the full validation output.
