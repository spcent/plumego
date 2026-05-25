# Card 2066

Milestone: M-024
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P2
State: blocked
Blocked By: card 2065 and M-022 merge; same middleware documentation surface
Primary Module: middleware
Owned Files:
- `middleware/conformance/README.md`
- `docs/modules/middleware/README.md`
Depends On: 2065

## Goal

Make `middleware/conformance` discoverable as the shared test-only conformance
suite for stable middleware packages.

## Scope

Add a local README at `middleware/conformance/` and cross-link it from the
middleware primer so the directory no longer looks like an abandoned
implementation skeleton.

## Non-goals

- Do not add runtime code to `middleware/conformance`.
- Do not change middleware behavior or test semantics.
- Do not introduce a new module manifest just for this test-only directory.

## Files

- `middleware/conformance/README.md`
- `docs/modules/middleware/README.md`

## Acceptance Tests

<!-- none; docs-only discoverability card -->

## Tests

- `go test -timeout 20s ./middleware/...`

## Docs Sync

- `docs/modules/middleware/README.md`

## Validation

- `go test -timeout 20s ./middleware/...`
- `gofmt -l .`

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

Blocked pending card 2065 and M-022 merge.
