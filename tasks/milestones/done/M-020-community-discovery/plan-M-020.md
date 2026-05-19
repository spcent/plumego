# Plan for M-020: Community Discovery & Documentation Overhaul

Milestone: `M-020`
Branch: `milestone/M-020-community-discovery`
Affected Modules: docs, stable root doc comments.

## Phase Map

1. Orient: audit README structure and package-level docs.
2. Implement: update README, README_CN, stable root doc comments, and runnable examples.
3. Validate: run example tests, vet, gofmt, and module manifest checks.
4. Ship: record verification and move the milestone after review.

## Card Inventory

- README overhaul.
- Stable root package comments and examples.
- Extension package comments where needed for docs parity.

## Dependency Edges

- README structure and examples can be edited independently.
- Validation requires all docs/example edits to be complete.

## Parallel Groups

- README and README_CN updates.
- Stable root doc comments.
- Example tests.

## Risk Register

- README and README_CN drift. Mitigation: keep sections mirrored.
- Example functions that compile but do not execute cleanly. Mitigation: run `go test -run=Example ./...`.
- Accidental behavior change in doc-only milestone. Mitigation: inspect diff for source changes outside comments/examples.

## Verification Strategy

- `go vet ./...`
- `go test -run=Example ./...`
- `gofmt -l .`
- `go run ./internal/checks/module-manifests`

## Exit Condition

M-020 is complete when the README pair is mirrored, stable root package docs are present, example tests pass, and verify-M-020 records the commands run.
