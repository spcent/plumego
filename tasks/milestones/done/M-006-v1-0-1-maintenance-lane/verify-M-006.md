# Verify M-006: v1.0.1 Maintenance Lane

Milestone: `M-006`
Branch: `main` local execution

Verified Cards:
- 1439: v1.0.1 maintenance control-plane plan
- 1440: generated release data sync
- 1441: GitHub Actions Node runtime maintenance
- 1442: CLI install and onboarding guidance cleanup
- 1443: post-v1 release evidence index

## Scope Check

- In-scope files touched:
  - `tasks/milestones/done/M-006-v1-0-1-maintenance-lane/M-006.md`
  - `tasks/milestones/done/M-006-v1-0-1-maintenance-lane/plan-M-006.md`
  - `tasks/milestones/done/M-006-v1-0-1-maintenance-lane/verify-M-006.md`
  - `tasks/cards/active/README.md`
  - M-006 task cards under `tasks/cards/`
  - `.github/workflows/quality-gates.yml`
  - `website/src/generated/releases.ts`
  - website dev-server docs
  - `docs/release/v1.0.0.md`
  - `docs/release/POST_V1_EVIDENCE.md`
- Out-of-scope files touched: none intentionally.

## Ownership Check

- overlapping card ownership: none recorded.
- unresolved ownership conflicts: none.

## Symbol Completeness Check

- exported symbol changes: none recorded.
- residual reference grep: not applicable; M-006 was a maintenance and evidence
  lane.

## Published v1 Facts

- Final tag: `v1.0.0`
- Tag object: `cde6de1d`
- Tag target: `6a99c5e0`
- Final tag GitHub Actions run: `25922384589`
- Final tag GitHub Actions result: PASS
- Rule: do not rewrite the published tag.

## Card Evidence Summary

| Card | Result | Evidence |
| --- | --- | --- |
| 1439 | PASS | M-006 milestone, plan, and cards created |
| 1440 | PASS | `cd website && pnpm sync` preserved generated release-data normalization |
| 1441 | PASS | quality-gates workflow actions updated to v6; remote run `25954419567` passed |
| 1442 | PASS | CLI install docs aligned to source-checkout path; CLI/docs/website checks passed |
| 1443 | PASS | `docs/release/POST_V1_EVIDENCE.md` added |

## Validation Summary

| Check | Result |
| --- | --- |
| `go run ./internal/checks/agent-workflow` | PASS |
| `go run ./internal/checks/module-manifests` | PASS in card 1439 |
| `go run ./internal/checks/extension-beta-evidence` | PASS |
| `bash scripts/check-doc-snippets-compile.sh` | PASS in card 1442 |
| `cd website && pnpm check` | PASS in card 1442 |
| `cd cmd/plumego && TMPDIR=/private/tmp GOCACHE=/private/tmp/plumego-gocache go test -timeout 60s ./...` | PASS in card 1442 |

## Module Test Summary

- primary module tests: CLI and website checks passed in card-level evidence.
- secondary module tests: extension evidence and maturity checks passed.

## Boundary Check Summary

- dependency-rules: not separately recorded in this verify artifact.
- agent-workflow: PASS.
- module-manifests: PASS in card 1439.
- reference-layout: not separately recorded in this verify artifact.
- public-entrypoints-sync: not required by the original M-006 scope.

## Repo Gate Summary

- `go test -race -timeout 60s ./...`: not part of the original M-006
  acceptance scope.
- `go test -timeout 20s ./...`: see Open Issues for local sandbox timeout and
  60s retry evidence.
- `go vet ./...`: not separately recorded in this verify artifact.
- `gofmt -l .`: covered by card-level validation.

## Extension Maturity Summary

Beta remains limited to:

- `x/gateway`
- `x/observability`
- `x/rest`
- `x/websocket`

Remaining candidates and surfaces still retain explicit evidence blockers in
`specs/extension-beta-evidence.yaml`.

## Open Issues

- `cd cmd/plumego && go test -timeout 20s ./...` was unreliable in the local
  sandbox because the default Go cache path was not writable and the retry hit
  the 20s timeout during temp cleanup. The same test set passed with writable
  temp/cache paths and a 60s timeout.
- Existing GitHub dependency alerts are outside M-006 scope.

## Final Verdict

- `PASS`
- rationale: post-v1 maintenance control plane is in place, generated data and
  onboarding drift are resolved, CI action runtime warnings are fixed, and
  extension promotion blockers remain explicit.
