# Verify M-017: gRPC Support

Milestone: `M-017`
Branch: `milestone/M-017-grpc-support`
Verified Cards: 1487, 1488, 1489

## Scope Check

- In-scope files touched: `x/rpc`, `reference/with-rpc`, and this verify artifact.
- Out-of-scope files touched: stable roots remain free of gRPC imports.

## Ownership Check

- overlapping card ownership: `reference/with-rpc` owns concrete gRPC and grpc-gateway dependencies.
- unresolved ownership conflicts: none.

## Symbol Completeness Check

- exported symbol changes: additive experimental `x/rpc` surface.
- residual reference grep: root `go.mod` has no gRPC dependency; gRPC dependencies are confined to `reference/with-rpc`.

## Acceptance Test Results

| Check | Result |
| --- | --- |
| `x/rpc/server`, `x/rpc/client`, `x/rpc/gateway` exist with tests | PASS |
| `reference/with-rpc` builds | PASS |
| `reference/with-rpc` tests | PASS |
| no gRPC dependency in main module `go.mod` | PASS |

## Module Test Summary

- primary module tests: `go test -timeout 60s ./x/rpc/...` PASS.
- reference tests: `go build ./...` and `go test -timeout 60s ./...` from `reference/with-rpc` PASS.

## Boundary Check Summary

- dependency-rules: PASS.
- agent-workflow: PASS.
- module-manifests: PASS.
- reference-layout: PASS.
- public-entrypoints-sync: not required; no stable public API change.

## Repo Gate Summary

- `go test -race -timeout 60s ./...`: not rerun in this cleanup pass.
- `go test -timeout 120s ./...`: PASS.
- `go vet ./...`: PASS.
- `gofmt -l .`: PASS.

## Checkpoint Summary

- Phase 1: existing `x/rpc` and reference wiring were inspected.
- Phase 2: reference dependency drift was fixed while preserving Go 1.26.
- Phase 3: focused x/rpc and reference/with-rpc build/test checks passed.

## Open Issues

- `x/rpc` is currently a transport-neutral wrapper; concrete gRPC dependencies live in `reference/with-rpc`. This is a deliberate boundary-preserving implementation choice and should be reflected in any future M-017 follow-up spec.

## Final Verdict

- `PASS`
- rationale: the implemented RPC surface and reference application build and test successfully, with concrete gRPC dependencies kept outside the main module.
