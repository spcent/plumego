# Plumego Benchmarks

This directory contains benchmark results comparing Plumego against Chi, Gin,
and Echo. Results are collected in CI and checked in after each release.

## How to Read These Numbers

Plumego preserves full `net/http` compatibility. Every request carries a
`context.WithValue` call and a `req.WithContext` allocation that Gin and Echo
avoid by using their own context types. The parallel numbers reflect this
accurately: Gin/Echo route ~2.5× faster per goroutine when measured in isolation
because they skip that allocation.

In practice, most services spend orders of magnitude more time on database
queries, serialisation, and downstream calls. The router's nanosecond overhead
rarely appears in production profiles.

## Latest Results (v1.1.0)

See [`results-v1.1.0.md`](./results-v1.1.0.md) for full raw output.

### Summary (medians, sequential dispatch)

| Scenario | Plumego | Chi | Gin | Echo |
|---|---|---|---|---|
| Static route (ns/op) | 4208 | 4110 | 3780 | 3721 |
| Single param (ns/op) | 4291 | 4547 | 3773 | 3743 |
| Multi-param (ns/op) | 4233 | 4705 | 3892 | 4023 |
| Scale-100 routes (ns/op) | 4342 | 4405 | 3981 | 5398 |
| Scale-500 routes (ns/op) | 6034 | 6258 | 5593 | 5508 |
| 404 not found (ns/op) | 6140 | 6596 | 5937 | 8608 |

### Summary (medians, parallel dispatch)

| Router | ns/op | allocs/op |
|---|---|---|
| Plumego | 663 | 7 |
| Chi | 1013 | 8 |
| Gin | 285 | 4 |
| Echo | 231 | 4 |

**Key takeaways:**

- Plumego is faster than Chi across all sequential scenarios.
- Plumego is ~12% slower than Gin/Echo in single-request sequential dispatch.
- The parallel gap (~2.5×) is entirely explained by `context.WithValue` +
  `req.WithContext`, which Gin/Echo eliminate by abandoning `net/http`
  handler shape.
- 404 handling: Plumego is faster than Echo and competitive with Chi.
- At 500 registered routes, Plumego's cache sustains near-flat performance.

## What the Gap Means

The `net/http` compatibility constraint costs ~3 allocs/op versus Gin/Echo.
That cost buys:

- Ordinary `func(http.ResponseWriter, *http.Request)` handlers — no framework
  context type anywhere in application code.
- Full compatibility with `httptest`, `net/http/httputil`, standard middleware,
  and any Go HTTP library that accepts `http.Handler`.
- Zero coupling to a framework-specific request lifecycle.

For most internal and platform APIs, the per-request framework overhead is
unmeasurable against real workloads. This trade-off is correct for services
where long-term maintainability and standard-library interop matter.

If raw throughput is the primary constraint (edge proxies, high-volume public
gateways), see [`docs/when-not-to-use-plumego.md`](../when-not-to-use-plumego.md).

## Middleware Chain Results

At v1.1.0, a 1-middleware chain with JSON response costs ~4.6 µs/op; a
5-middleware chain costs ~5.2 µs/op. Gin comparably scales from ~5.1 µs/op
to ~5.9 µs/op. The chains are within measurement noise of each other.

## Reproducing Results

```bash
cd reference/benchmark
go test -bench=. -benchmem -count=3 ./...
```

Results are collected on a dedicated CI runner. Consumer hardware numbers will
differ. Compare ratios, not absolute values.

## Running Router-Only Benchmarks

```bash
go test -bench=BenchmarkOpt -benchmem -count=3 ./router/
```

This covers static, single-param, multi-param, wildcard, many-route scale,
deep-path, parallel, and path-normalization scenarios without the full HTTP
stack overhead.
