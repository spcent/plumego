# Verify M-013: Migration Guides

Milestone: `M-013`
Branch:
Verified Cards:

## Scope Check

- In-scope files touched:
- Out-of-scope files touched:

## Ownership Check

- overlapping card ownership:
- unresolved ownership conflicts:

## Symbol Completeness Check

- exported symbol changes:
- residual reference grep:

## Acceptance Test Results

<!-- List each acceptance test defined in task cards and its result.
     Format: <file>: <TestFunctionName>: PASS / FAIL -->

## Module Test Summary

- primary module tests:
- secondary module tests:

## Boundary Check Summary

- dependency-rules:
- agent-workflow:
- module-manifests:
- reference-layout:
- public-entrypoints-sync:

## Repo Gate Summary

- `go test -race -timeout 60s ./...`
- `go test -timeout 20s ./...`
- `go vet ./...`
- `gofmt -l .`

## Checkpoint Summary

<!-- Read checkpoint-M-013.json from this milestone directory and summarize.
     Or run: make milestone-status M=active/M-013-short-name -->

- Phase 1:
- Phase 2:
- Phase 3:

## Open Issues

- none

## Final Verdict

- `PASS` or `FAIL`
- rationale:
