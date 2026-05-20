# Verify M-009: Beta Promotions Round 1

Milestone: `M-009`
Branch: `milestone/M-009-beta-promotions-r1`
Verified Cards: 1367, 1370, 1371, 1372, 1513

## Scope Check

- In-scope files touched: extension evidence ledger, extension maturity docs, deprecation docs, and promoted module manifests.
- Out-of-scope files touched: runtime behavior changes were out of scope.

## Ownership Check

- overlapping card ownership: extension evidence and maturity dashboard updates overlapped intentionally.
- unresolved ownership conflicts: `x/gateway/discovery` promotion remains blocked.

## Symbol Completeness Check

- exported symbol changes: none expected; evidence-only promotions.
- residual reference grep: not applicable.

## Acceptance Test Results

| Surface | Result |
| --- | --- |
| `x/ai/provider`, `x/ai/session`, `x/ai/streaming`, `x/ai/tool` beta evidence | PASS |
| `x/tenant` beta evidence | PASS |
| `x/data/file` and `x/data/idempotency` beta surface evidence | PASS |
| `x/gateway/discovery:core-static` beta evidence | FAIL, still blocked |

## Module Test Summary

- primary module tests: `go run ./internal/checks/extension-beta-evidence` reports all non-discovery M-009 surfaces clear.
- secondary module tests: not required; no runtime changes.

## Boundary Check Summary

- dependency-rules: pending final staged validation.
- agent-workflow: pending final staged validation.
- module-manifests: pending final staged validation.
- reference-layout: not required.
- public-entrypoints-sync: not required.

## Repo Gate Summary

- `go test -race -timeout 60s ./...`: not run for this evidence-only pass.
- `go test -timeout 20s ./...`: not run for this evidence-only pass.
- `go vet ./...`: not run for this evidence-only pass.
- `gofmt -l .`: pending final staged validation.

## Checkpoint Summary

- Phase 1: v1.1.0 release refs were confirmed.
- Phase 2: eligible M-009 surfaces were promoted.
- Phase 3: dashboard docs were updated with discovery kept blocked.

## Open Issues

- `x/gateway/discovery:core-static` remains blocked on API snapshot and owner sign-off evidence.

## Final Verdict

- `FAIL`
- rationale: most M-009 promotions are complete, but one required surface remains explicitly blocked.
