# Verify M-001: v1 Trust Baseline

Milestone: `M-001`
Branch: `main`
Verified Cards: historical v1 trust-baseline and CLI scaffold work recorded in
`tasks/cards/done/`

## Scope Check

- In-scope files touched by this verify pass: this verify artifact only.
- Out-of-scope files touched: none.

## Ownership Check

- overlapping card ownership: card 1429 owns this verification artifact.
- unresolved ownership conflicts: none.

## Symbol Completeness Check

- exported symbol changes: none.
- residual reference grep: not applicable.

## Template and Release Truth Matrix

| Item | Evidence | Result |
| --- | --- | --- |
| Release tag status | `git tag -l 'v*'` returned `v0.2.0`; `v1.0.0-rc.1` is not present locally | WARN: rc tag still blocked by card 1430 |
| CLI module tests | `cd cmd/plumego && go test -timeout 20s ./...` | PASS |
| Canonical dry-run | `cd cmd/plumego && go run . new --template canonical --dry-run trust-check` | PASS |
| REST dry-run | `cd cmd/plumego && go run . new --template rest-api --dry-run trust-check` | PASS |
| Invalid template | `cd cmd/plumego && go run . new --template invalid-template-name --dry-run trust-check` returned exit status 3 with valid-template list | PASS |

## Module Test Summary

- primary module tests: `cmd/plumego` submodule tests passed.
- secondary module tests: not required for this trust-baseline verify pass.

## Boundary Check Summary

- dependency-rules: not run in this verify pass.
- agent-workflow: PASS.
- module-manifests: PASS.
- reference-layout: not run in this verify pass.

## Repo Gate Summary

- `go test -race -timeout 60s ./...`: not run in this verify pass.
- `go test -timeout 20s ./...`: not run repo-wide; `cmd/plumego` submodule tests passed.
- `go vet ./...`: not run in this verify pass.
- `gofmt -l .`: not run in this verify pass.

## Open Issues

- `v1.0.0-rc.1` is still not tagged. This is tracked by card 1430 and existing
  card 1375.

## Final Verdict

- `PASS-WITH-WARN`
- rationale: CLI scaffold trust evidence is current and executable; the release
  tag remains an explicit downstream blocker rather than hidden state.
