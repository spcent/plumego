# 0676 - router pattern cache buffer completeness

State: active
Priority: P0
Primary module: `router`

## Goal

Remove fixed-buffer truncation from parameterized route pattern-cache matching
so cached wildcard and high-parameter routes preserve full path and parameter
values.

## Scope

- Change pattern-cache path splitting to grow beyond the pooled initial
  capacity.
- Change cached parameter extraction to preserve more than the initial pooled
  parameter capacity.
- Add regression tests for cached wildcard routes with long remainders and
  cached routes with many parameters.

## Non-goals

- Do not replace the trie matcher or change route precedence.
- Do not expose cache tuning as public API.
- Do not add non-standard-library dependencies.

## Files

- `router/path.go`
- `router/cache.go`
- `router/cache_coverage_test.go`
- `router/router_contract_test.go`

## Tests

- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`

## Docs Sync

No docs sync expected; this is internal cache correctness.

## Done Definition

- Pattern cache matching does not truncate paths beyond the initial segment
  buffer.
- Pattern cache parameter extraction does not drop params beyond the initial
  parameter buffer.
- Router validation commands pass.

## Outcome

