# Card 0321: Metrics Record Buffer Boundary Cleanup

Priority: P1
State: done
Primary Module: metrics

## Goal

Remove record-buffer retention and inspection behavior from the stable `metrics` root so `BaseMetricsCollector` matches the documented stable boundary.

## Problem

- `metrics/collector.go` still stores full metric records in memory via:
  - `records []MetricRecord`
  - `maxRecords`
  - `setMaxRecords(...)`
  - `recordsSnapshot()`
- This directly conflicts with the metrics control plane:
  - `metrics/module.yaml` says record-buffer helpers and retention policy do not belong in stable collectors.
  - `docs/modules/metrics/README.md` says record-buffer inspection and retention tuning belong in `x/observability/recordbuffer`.
- The stable root should expose aggregate collector contracts and stats, not a hidden in-memory record log that tests and future callers can grow into a quasi-feature surface.

## Scope

- Remove stable record-buffer state and helpers from `BaseMetricsCollector`.
- Rewrite `metrics/collector_test.go` to assert on stable behavior through `GetStats()` and externally visible collector effects instead of snapshotting internal record buffers.
- Keep the public stable surface limited to `Record`, `ObserveHTTP`, `GetStats`, and `Clear`.

## Non-Goals

- Do not add a replacement record-buffer package in this card.
- Do not redesign `MetricRecord`, `CollectorStats`, or `HTTPObserver`.
- Do not widen `metrics` into a richer in-memory analytics surface.

## Files

- `metrics/collector.go`
- `metrics/collector_test.go`
- `docs/modules/metrics/README.md`
- `metrics/TESTING.md`

## Tests

- `go test -timeout 20s ./metrics/...`
- `go test -race -timeout 60s ./metrics/...`
- `go vet ./metrics/...`

## Docs Sync

- Sync `docs/modules/metrics/README.md` and `metrics/TESTING.md` if the stable behavior or recommended landing zone needs sharper wording after the cleanup.

## Done Definition

- Stable `metrics` no longer keeps an internal retained record buffer.
- Tests stop depending on `recordsSnapshot()` or `setMaxRecords(...)`.
- `BaseMetricsCollector` aligns with the documented “stats and shared HTTP observation only” stable role.
- The testing guide no longer teaches extension-only observer helpers as part of the stable collector surface.

## Outcome

- Removed retained per-record buffer state from `BaseMetricsCollector`.
- Deleted the package-local record-buffer helpers and rewrote the tests to assert only stable collector behavior through `GetStats()` and `Clear()`.
- Tightened the metrics module docs so they explicitly forbid stable per-record retention and point record inspection to `x/observability/recordbuffer`.
- Rewrote `metrics/TESTING.md` so it only documents the stable collector contract and points richer feature-specific test helpers back to `x/observability`.

## Validation Run

```bash
gofmt -w metrics/collector.go metrics/collector_test.go
go test -timeout 20s ./metrics/...
go test -race -timeout 60s ./metrics/...
go vet ./metrics/...
```
