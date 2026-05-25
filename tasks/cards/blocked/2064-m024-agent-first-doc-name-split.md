# Card 2064

Milestone: M-024
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: control-plane
Priority: P1
State: blocked
Blocked By: M-022 merge; overlaps docs/control-plane follow-up work
Primary Module: docs
Owned Files:
- `AGENTS.md`
- `docs/README.md`
- `docs/AGENT_FIRST.md`
- `docs/agent-first.md`
- `README.md`
Depends On: M-022

## Goal

Replace the case-only `docs/AGENT_FIRST.md` vs `docs/agent-first.md` split with
clearly distinguished internal vs external documentation names and references.

## Scope

Rename or reframe the internal document path, keep the external-facing guide
from M-021 intact in purpose, and update all first-read indexes and authority
links so agents do not have to infer intent from filename case.

## Non-goals

- Do not rewrite the substantive content of the external M-021 guide beyond
  naming and authority clarification.
- Do not turn this card into a broad docs architecture rewrite.
- Do not mutate historical milestone outcome prose except where a link must be updated.

## Files

- `AGENTS.md`
- `docs/README.md`
- `docs/AGENT_FIRST.md`
- `docs/agent-first.md`
- `README.md`

## Acceptance Tests

<!-- none; naming and link-clarity card -->

## Tests

- `go run ./internal/checks/agent-workflow`

## Docs Sync

- `AGENTS.md`
- `docs/README.md`
- `README.md`

## Validation

- `go run ./internal/checks/agent-workflow`
- `gofmt -l .`

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

Blocked pending M-022 merge.
