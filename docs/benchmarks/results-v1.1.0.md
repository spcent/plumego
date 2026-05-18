# Plumego v1.1.0 Benchmark Results

Date: 2026-05-18T16:12:26Z
Go toolchain: go version go1.26.0 darwin/arm64
OS/arch: Darwin 25.3.0 arm64
CPU: Apple M1 Pro
Command: `GOCACHE=/private/tmp/plumego-gocache go test -bench=. -benchmem -count=3 ./...`
Working directory: `benchmark/`

The output below is verbatim stdout/stderr from the benchmark command.

```text
goos: darwin
goarch: arm64
pkg: github.com/spcent/plumego/benchmark
cpu: Apple M1 Pro
BenchmarkChain1NoOp-10           	  836576	      1528 ns/op	    6058 B/op	      17 allocs/op
BenchmarkChain1NoOp-10           	  748076	      1670 ns/op	    6058 B/op	      17 allocs/op
BenchmarkChain1NoOp-10           	  760011	      1486 ns/op	    6058 B/op	      17 allocs/op
BenchmarkChain3NoOp-10           	  748846	      1668 ns/op	    6186 B/op	      19 allocs/op
BenchmarkChain3NoOp-10           	  733987	      1615 ns/op	    6186 B/op	      19 allocs/op
BenchmarkChain3NoOp-10           	  798621	      1620 ns/op	    6186 B/op	      19 allocs/op
BenchmarkChain5NoOp-10           	  698940	      1795 ns/op	    6346 B/op	      20 allocs/op
BenchmarkChain5NoOp-10           	  637216	      1881 ns/op	    6346 B/op	      20 allocs/op
BenchmarkChain5NoOp-10           	  635254	      1970 ns/op	    6346 B/op	      20 allocs/op
BenchmarkChain1JSON-10           	  625904	      1880 ns/op	    6285 B/op	      21 allocs/op
BenchmarkChain1JSON-10           	  676885	      2006 ns/op	    6285 B/op	      21 allocs/op
BenchmarkChain1JSON-10           	  669976	      1930 ns/op	    6285 B/op	      21 allocs/op
BenchmarkChain3JSON-10           	  538282	      2051 ns/op	    6413 B/op	      23 allocs/op
BenchmarkChain3JSON-10           	  576226	      1990 ns/op	    6413 B/op	      23 allocs/op
BenchmarkChain3JSON-10           	  562508	      2006 ns/op	    6413 B/op	      23 allocs/op
BenchmarkChain5JSON-10           	  535617	      2048 ns/op	    6573 B/op	      24 allocs/op
BenchmarkChain5JSON-10           	  606829	      2045 ns/op	    6573 B/op	      24 allocs/op
BenchmarkChain5JSON-10           	  558681	      2042 ns/op	    6573 B/op	      24 allocs/op
BenchmarkRouterStatic/plumego-10 	  922920	      1302 ns/op	    5818 B/op	      18 allocs/op
BenchmarkRouterStatic/plumego-10 	  873435	      1392 ns/op	    5818 B/op	      18 allocs/op
BenchmarkRouterStatic/plumego-10 	  926500	      1443 ns/op	    5818 B/op	      18 allocs/op
BenchmarkRouterStatic/chi-10     	  884541	      1441 ns/op	    5701 B/op	      15 allocs/op
BenchmarkRouterStatic/chi-10     	  761868	      1382 ns/op	    5701 B/op	      15 allocs/op
BenchmarkRouterStatic/chi-10     	  835555	      1455 ns/op	    5701 B/op	      15 allocs/op
BenchmarkRouterSingleParam/plumego-10         	  689760	      1621 ns/op	    6434 B/op	      21 allocs/op
BenchmarkRouterSingleParam/plumego-10         	  680506	      1622 ns/op	    6434 B/op	      21 allocs/op
BenchmarkRouterSingleParam/plumego-10         	  789206	      1726 ns/op	    6434 B/op	      21 allocs/op
BenchmarkRouterSingleParam/chi-10             	  765441	      1622 ns/op	    6037 B/op	      17 allocs/op
BenchmarkRouterSingleParam/chi-10             	  728232	      1548 ns/op	    6037 B/op	      17 allocs/op
BenchmarkRouterSingleParam/chi-10             	  714044	      1534 ns/op	    6037 B/op	      17 allocs/op
BenchmarkRouterMultiParam/plumego-10          	  594865	      1949 ns/op	    6562 B/op	      22 allocs/op
BenchmarkRouterMultiParam/plumego-10          	  653127	      1763 ns/op	    6562 B/op	      22 allocs/op
BenchmarkRouterMultiParam/plumego-10          	  686961	      1719 ns/op	    6562 B/op	      22 allocs/op
BenchmarkRouterMultiParam/chi-10              	  748471	      1964 ns/op	    6086 B/op	      17 allocs/op
BenchmarkRouterMultiParam/chi-10              	  659936	      1788 ns/op	    6085 B/op	      17 allocs/op
BenchmarkRouterMultiParam/chi-10              	  800208	      1897 ns/op	    6085 B/op	      17 allocs/op
PASS
ok  	github.com/spcent/plumego/benchmark	45.059s
```
