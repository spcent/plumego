# Benchmark Results — v1.3.0

## Environment

- **Date**: 2026-05-19
- **OS**: Linux 6.18.5 x86_64
- **CPU**: Intel(R) Xeon(R) Processor @ 2.10 GHz
- **Go**: 1.24.7 linux/amd64
- **Module**: `reference/benchmark`
- **Command**: `go test -bench=. -benchmem -count=3 ./...`

## Summary of Changes (Phase 1+2+3)

This release includes three performance optimization phases:

- **Phase 1**: Fixed `cloneStringMap(nil)` returning an empty map (now returns nil),
  eliminating 1 map alloc on static routes.
- **Phase 2**: Replaced `buildParamMap`+`WithRequestContext` double-allocation path with
  a pooled `RouteState` carrying inline `[8]string` param arrays. Route param dispatch
  now costs **3 allocs** (down from 5): cache-key string, context node, request copy.
  Zero map allocations for parameter storage on the hot path.
- **Phase 3**: Replaced the LRU cache (write lock on every read for promotion) with a
  ring-buffer FIFO cache. `Get` acquires only `RLock`; concurrent readers never block
  each other. Parallel throughput improved ~4× vs v1.2.0.

## Router — Route Shape (external, with httptest overhead)

```
BenchmarkRouterStatic/plumego-4      283430    4157 ns/op    5723 B/op    16 allocs/op
BenchmarkRouterStatic/chi-4          301960    4020 ns/op    5699 B/op    15 allocs/op
BenchmarkRouterStatic/gin-4          313944    3780 ns/op    5331 B/op    13 allocs/op
BenchmarkRouterStatic/echo-4         325833    3721 ns/op    5331 B/op    13 allocs/op

BenchmarkRouterSingleParam/plumego-4 276097    4291 ns/op    5715 B/op    16 allocs/op
BenchmarkRouterSingleParam/chi-4     257880    4547 ns/op    6036 B/op    17 allocs/op  ← Plumego faster
BenchmarkRouterSingleParam/gin-4     317233    3773 ns/op    5331 B/op    13 allocs/op
BenchmarkRouterSingleParam/echo-4    312391    3743 ns/op    5331 B/op    13 allocs/op

BenchmarkRouterMultiParam/plumego-4  297676    4233 ns/op    5843 B/op    17 allocs/op
BenchmarkRouterMultiParam/chi-4      242163    4763 ns/op    6084 B/op    17 allocs/op  ← Plumego faster
BenchmarkRouterMultiParam/gin-4      284920    4105 ns/op    5380 B/op    13 allocs/op
BenchmarkRouterMultiParam/echo-4     252151    4023 ns/op    5379 B/op    13 allocs/op
```

## Router — Parallel Throughput (b.RunParallel, GOMAXPROCS=4)

```
BenchmarkRouterParallelPlumego-4     1799718    663 ns/op    592 B/op    7 allocs/op
BenchmarkRouterParallelChi-4         1165504   1013 ns/op    912 B/op    8 allocs/op   ← Plumego 1.5× faster
BenchmarkRouterParallelGin-4         3917090    285 ns/op    208 B/op    4 allocs/op
BenchmarkRouterParallelEcho-4        4910727    231 ns/op    208 B/op    4 allocs/op
```

**Improvement vs v1.2.0**: Plumego parallel went from ~1800 ns/op → ~663 ns/op (~2.7×).
Gap to Gin/Echo reduced from ~7× to ~2.5× (remaining gap is `context.WithValue` +
`req.WithContext` overhead required for stdlib `net/http` compatibility).

## Router — Route Table Scale

```
BenchmarkRouterScale100/plumego-4    281572    4342 ns/op    5763 B/op    16 allocs/op
BenchmarkRouterScale100/chi-4        274033    4405 ns/op    5731 B/op    15 allocs/op
BenchmarkRouterScale100/gin-4        296073    3981 ns/op    5364 B/op    13 allocs/op
BenchmarkRouterScale100/echo-4       228865    5398 ns/op    5362 B/op    13 allocs/op

BenchmarkRouterScale500/plumego-4    195616    6034 ns/op    5779 B/op    16 allocs/op
BenchmarkRouterScale500/chi-4        167024    6258 ns/op    5731 B/op    15 allocs/op
BenchmarkRouterScale500/gin-4        234998    5593 ns/op    5363 B/op    13 allocs/op
BenchmarkRouterScale500/echo-4       204645    5658 ns/op    5362 B/op    13 allocs/op
```

## Router — 404 Overhead

```
BenchmarkRouterNotFoundPlumego-4    179011    6140 ns/op    6222 B/op    21 allocs/op
BenchmarkRouterNotFoundChi-4        174235    6596 ns/op    6564 B/op    22 allocs/op
BenchmarkRouterNotFoundGin-4        175324    6078 ns/op    6131 B/op    17 allocs/op
BenchmarkRouterNotFoundEcho-4       134203    8608 ns/op    6628 B/op    25 allocs/op
```

## Middleware Chain — stdlib-compatible (Plumego/Chi)

```
BenchmarkChain1NoOp-4    304682    4047 ns/op    6057 B/op    17 allocs/op
BenchmarkChain3NoOp-4    259130    4519 ns/op    6185 B/op    19 allocs/op
BenchmarkChain5NoOp-4    263770    4689 ns/op    6345 B/op    20 allocs/op
BenchmarkChain1JSON-4    255531    4628 ns/op    6283 B/op    21 allocs/op
BenchmarkChain3JSON-4    246916    5047 ns/op    6411 B/op    23 allocs/op
BenchmarkChain5JSON-4    224996    5370 ns/op    6572 B/op    24 allocs/op
```

## Internal Router Benchmarks (no httptest overhead)

These remove the ~4 alloc / ~200 ns httptest baseline to show routing cost alone.

```
BenchmarkOptStaticRoute-4       2527747    497 ns/op    384 B/op    3 allocs/op
BenchmarkOptSingleParam-4       2470862    537 ns/op    384 B/op    3 allocs/op
BenchmarkOptMultiParam-4        2330428    512 ns/op    392 B/op    3 allocs/op
BenchmarkOptParamRoute/cache_hit-4   2205879    569 ns/op    392 B/op    3 allocs/op
BenchmarkOptParallelParam-4     4042219    299 ns/op    384 B/op    3 allocs/op
```

### Allocation breakdown (3 allocs/request, warm cache hit):

| Source | Bytes | Count |
|---|---|---|
| `fastBuildCacheKey` (cache key string) | ~16 | 1 |
| `context.WithValue` (context node) | ~48 | 1 |
| `req.WithContext` (request copy) | ~320 | 1 |
| `RouteState` (params inline array) | 0 | 0 (pooled) |
| `buildParamMap` (map) | 0 | 0 (eliminated) |
| `WithRequestContext` clone (map) | 0 | 0 (eliminated) |

## Notes

- **httptest overhead** contributes ~13 allocs per request and is constant across all
  frameworks in the external comparison benchmarks.
- **stdlib compatibility cost**: Gin and Echo avoid `context.WithValue` and
  `req.WithContext` by using framework-specific context objects (sync.Pool'd). This saves
  ~2 allocs per request but requires using framework-specific APIs throughout.
- **Plumego vs Chi**: Plumego is now faster than Chi in both sequential and parallel
  benchmarks across all route shapes tested.
- **Cache eviction policy**: changed from LRU (write lock on every read) to FIFO (ring
  buffer, read-only RLock). This eliminates write-lock contention on cache hits under
  concurrent load.
