# 1512 — x/ai/metrics interface convergence decision

## Context

`x/ai/metrics.Collector` and `metrics.AggregateCollector` (stable root) are two
parallel interfaces bridged by `AggregateCollectorAdapter`. The x/ai/metrics package
intentionally uses a Counter/Gauge/Histogram/Timing method set optimised for AI
provider telemetry (model name, token counts, latency tags), whereas the stable
`metrics.AggregateCollector` uses a single `Record(ctx, MetricRecord)` model suited
for general HTTP and app-level metrics.

## Decision required

**Option A — Converge**: extend `metrics.AggregateCollector` with an AI-oriented
tag model so x/ai can use the stable interface directly and retire
`x/ai/metrics.Collector`.

Implications: stable root API change (needs M+ version bump), migration of all
x/ai metric call sites, removal of `AggregateCollectorAdapter`.

**Option B — Keep separate**: treat `x/ai/metrics.Collector` as a permanent
AI-scoped boundary. Freeze its interface, document `AggregateCollectorAdapter`
as the only bridge to the stable layer, and prohibit further divergence.

Implications: two collector interfaces coexist forever; clearer ownership but
higher cognitive load for developers working across AI and HTTP metrics.

## Acceptance criteria

- Decision recorded here with rationale
- If Option A: all x/ai metric call sites migrated, `x/ai/metrics.Collector` removed,
  `AggregateCollectorAdapter` removed, x/ai/metrics package retired or repurposed
- If Option B: `x/ai/metrics.Collector` interface frozen (no new methods),
  package comment updated to state the permanent boundary, `AggregateCollectorAdapter`
  documented as the only integration path

## Status

Open — pending discussion.
