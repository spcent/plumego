# Plumego v1.2.0 Benchmark Results

Date: 2026-05-19T00:00:00Z
Go toolchain: go version go1.24.7 linux/amd64
OS/arch: Linux 6.18.5 amd64
CPU: Intel(R) Xeon(R) Processor @ 2.80GHz
Command: `go test -bench=. -benchmem -count=3 ./...`
Working directory: `reference/benchmark/`
Frameworks: Plumego, Chi v5.2.5, Gin v1.10.0, Echo v4.13.3

The output below is verbatim stdout from the benchmark command.

## Key findings

**Sequential dispatch (single path-param route, medians):**

| Framework | ns/op | allocs/op | B/op | vs Plumego |
|---|---|---|---|---|
| Plumego | 4832 | 21 | 6433 | 1.0x |
| Chi | 4734 | 17 | 6035 | 0.98x |
| Gin | 4169 | 13 | 5331 | 0.86x |
| Echo | 3861 | 13 | 5331 | 0.80x |

**Parallel throughput (GOMAXPROCS goroutines, medians):**

| Framework | ns/op | allocs/op | B/op | vs Plumego |
|---|---|---|---|---|
| Plumego | 1107 | 12 | 1312 | 1.0x |
| Chi | 823 | 8 | 912 | 0.74x |
| Gin | 220 | 4 | 208 | 0.20x |
| Echo | 226 | 4 | 208 | 0.20x |

**Route table scale (single-param dispatch, medians):**

| Framework | 1 route | 100 routes | 500 routes |
|---|---|---|---|
| Plumego | 4832 ns | 5318 ns | 5608 ns |
| Chi | 4734 ns | 4452 ns | 4794 ns |
| Gin | 4169 ns | 4116 ns | 4335 ns |
| Echo | 3861 ns | 4264 ns | 4471 ns |

**Notes:**
- All sequential ns/op numbers include the full `httptest.NewRequest` + `ServeHTTP` + `httptest.NewRecorder` cycle.
- The ~3–4 µs httptest baseline is constant across all frameworks; relative numbers reflect true dispatch differences.
- Net per-request allocation for path params (single-param minus static baseline): Plumego ~616 B / 3 allocs, Chi ~336 B / 2 allocs, Gin 0 / 0 (pooled), Echo 0 / 0 (pooled).
- Gin and Echo use sync.Pool for context objects, which eliminates per-request context allocation. Under concurrent load this produces a ~5x throughput advantage.

## Raw output

```text
goos: linux
goarch: amd64
pkg: benchmark
cpu: Intel(R) Xeon(R) Processor @ 2.80GHz
BenchmarkChain1NoOp-4              	  217333	      5435 ns/op	    6057 B/op	      17 allocs/op
BenchmarkChain1NoOp-4              	  278979	      4473 ns/op	    6057 B/op	      17 allocs/op
BenchmarkChain1NoOp-4              	  264064	      4449 ns/op	    6057 B/op	      17 allocs/op
BenchmarkChain3NoOp-4              	  244974	      4755 ns/op	    6185 B/op	      19 allocs/op
BenchmarkChain3NoOp-4              	  255016	      4724 ns/op	    6185 B/op	      19 allocs/op
BenchmarkChain3NoOp-4              	  259556	      4719 ns/op	    6185 B/op	      19 allocs/op
BenchmarkChain5NoOp-4              	  232581	      5056 ns/op	    6345 B/op	      20 allocs/op
BenchmarkChain5NoOp-4              	  229975	      5169 ns/op	    6345 B/op	      20 allocs/op
BenchmarkChain5NoOp-4              	  216171	      6655 ns/op	    6345 B/op	      20 allocs/op
BenchmarkChain1JSON-4              	  164038	      7065 ns/op	    6283 B/op	      21 allocs/op
BenchmarkChain1JSON-4              	  211892	      5427 ns/op	    6283 B/op	      21 allocs/op
BenchmarkChain1JSON-4              	  220454	      5187 ns/op	    6283 B/op	      21 allocs/op
BenchmarkChain3JSON-4              	  218538	      5577 ns/op	    6411 B/op	      23 allocs/op
BenchmarkChain3JSON-4              	  213888	      5730 ns/op	    6411 B/op	      23 allocs/op
BenchmarkChain3JSON-4              	  204141	      5660 ns/op	    6411 B/op	      23 allocs/op
BenchmarkChain5JSON-4              	  194817	      7780 ns/op	    6572 B/op	      24 allocs/op
BenchmarkChain5JSON-4              	  156145	      7834 ns/op	    6572 B/op	      24 allocs/op
BenchmarkChain5JSON-4              	  146515	      7715 ns/op	    6572 B/op	      24 allocs/op
BenchmarkGinChain1NoOp-4           	  182910	      6072 ns/op	    6059 B/op	      17 allocs/op
BenchmarkGinChain1NoOp-4           	  231483	      5983 ns/op	    6060 B/op	      17 allocs/op
BenchmarkGinChain1NoOp-4           	  199485	      5943 ns/op	    6059 B/op	      17 allocs/op
BenchmarkGinChain3NoOp-4           	  194762	      6081 ns/op	    6092 B/op	      19 allocs/op
BenchmarkGinChain3NoOp-4           	  193594	      6185 ns/op	    6092 B/op	      19 allocs/op
BenchmarkGinChain3NoOp-4           	  195021	      6092 ns/op	    6092 B/op	      19 allocs/op
BenchmarkGinChain5NoOp-4           	  185292	      6409 ns/op	    6124 B/op	      21 allocs/op
BenchmarkGinChain5NoOp-4           	  193609	      6476 ns/op	    6124 B/op	      21 allocs/op
BenchmarkGinChain5NoOp-4           	  192012	      6319 ns/op	    6124 B/op	      21 allocs/op
BenchmarkGinChain1JSON-4           	  173301	      6841 ns/op	    6334 B/op	      25 allocs/op
BenchmarkGinChain1JSON-4           	  218792	      5908 ns/op	    6334 B/op	      25 allocs/op
BenchmarkGinChain1JSON-4           	  211458	      5991 ns/op	    6334 B/op	      25 allocs/op
BenchmarkGinChain3JSON-4           	  190810	      6258 ns/op	    6366 B/op	      27 allocs/op
BenchmarkGinChain3JSON-4           	  200037	      6227 ns/op	    6366 B/op	      27 allocs/op
BenchmarkGinChain3JSON-4           	  191539	      6093 ns/op	    6366 B/op	      27 allocs/op
BenchmarkGinChain5JSON-4           	  189362	      6616 ns/op	    6398 B/op	      29 allocs/op
BenchmarkGinChain5JSON-4           	  184180	      6530 ns/op	    6398 B/op	      29 allocs/op
BenchmarkGinChain5JSON-4           	  178969	      7929 ns/op	    6398 B/op	      29 allocs/op
BenchmarkEchoChain1NoOp-4          	  214393	      6164 ns/op	    6092 B/op	      18 allocs/op
BenchmarkEchoChain1NoOp-4          	  188826	      6261 ns/op	    6092 B/op	      18 allocs/op
BenchmarkEchoChain1NoOp-4          	  196616	      6270 ns/op	    6092 B/op	      18 allocs/op
BenchmarkEchoChain3NoOp-4          	  181464	      6678 ns/op	    6284 B/op	      22 allocs/op
BenchmarkEchoChain3NoOp-4          	  211929	      6380 ns/op	    6284 B/op	      22 allocs/op
BenchmarkEchoChain3NoOp-4          	  187011	      6652 ns/op	    6284 B/op	      22 allocs/op
BenchmarkEchoChain5NoOp-4          	  178869	      7271 ns/op	    6508 B/op	      25 allocs/op
BenchmarkEchoChain5NoOp-4          	  164113	      7474 ns/op	    6508 B/op	      25 allocs/op
BenchmarkEchoChain5NoOp-4          	  152002	      7575 ns/op	    6508 B/op	      25 allocs/op
BenchmarkEchoChain1JSON-4          	  151828	      7989 ns/op	    6382 B/op	      26 allocs/op
BenchmarkEchoChain1JSON-4          	  150112	      7492 ns/op	    6382 B/op	      26 allocs/op
BenchmarkEchoChain1JSON-4          	  159960	      7309 ns/op	    6382 B/op	      26 allocs/op
BenchmarkEchoChain3JSON-4          	  186931	      7883 ns/op	    6574 B/op	      30 allocs/op
BenchmarkEchoChain3JSON-4          	  129279	      9037 ns/op	    6574 B/op	      30 allocs/op
BenchmarkEchoChain3JSON-4          	  177408	      6142 ns/op	    6574 B/op	      30 allocs/op
BenchmarkEchoChain5JSON-4          	  191774	      6755 ns/op	    6798 B/op	      33 allocs/op
BenchmarkEchoChain5JSON-4          	  176178	      6645 ns/op	    6798 B/op	      33 allocs/op
BenchmarkEchoChain5JSON-4          	  189414	      7849 ns/op	    6798 B/op	      33 allocs/op
BenchmarkRouterStatic/plumego-4    	  199416	      6093 ns/op	    5817 B/op	      18 allocs/op
BenchmarkRouterStatic/plumego-4    	  198657	      6243 ns/op	    5817 B/op	      18 allocs/op
BenchmarkRouterStatic/plumego-4    	  205982	      6118 ns/op	    5817 B/op	      18 allocs/op
BenchmarkRouterStatic/chi-4        	  204988	      5124 ns/op	    5699 B/op	      15 allocs/op
BenchmarkRouterStatic/chi-4        	  236065	      4241 ns/op	    5699 B/op	      15 allocs/op
BenchmarkRouterStatic/chi-4        	  302598	      4304 ns/op	    5699 B/op	      15 allocs/op
BenchmarkRouterStatic/gin-4        	  292342	      5069 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/gin-4        	  233119	      5249 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/gin-4        	  316141	      3865 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/echo-4       	  296494	      3800 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/echo-4       	  318992	      3765 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/echo-4       	  336468	      3770 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/plumego-4         	  231216	      4982 ns/op	    6433 B/op	      21 allocs/op
BenchmarkRouterSingleParam/plumego-4         	  255394	      4797 ns/op	    6433 B/op	      21 allocs/op
BenchmarkRouterSingleParam/plumego-4         	  256585	      4832 ns/op	    6433 B/op	      21 allocs/op
BenchmarkRouterSingleParam/chi-4             	  256452	      4580 ns/op	    6035 B/op	      17 allocs/op
BenchmarkRouterSingleParam/chi-4             	  240411	      4734 ns/op	    6035 B/op	      17 allocs/op
BenchmarkRouterSingleParam/chi-4             	  247940	      4757 ns/op	    6035 B/op	      17 allocs/op
BenchmarkRouterSingleParam/gin-4             	  295269	      4169 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/gin-4             	  227133	      5047 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/gin-4             	  288147	      4010 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/echo-4            	  305361	      3836 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/echo-4            	  328160	      3861 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/echo-4            	  281694	      3966 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterMultiParam/plumego-4          	  230353	      5204 ns/op	    6561 B/op	      22 allocs/op
BenchmarkRouterMultiParam/plumego-4          	  225646	      6046 ns/op	    6561 B/op	      22 allocs/op
BenchmarkRouterMultiParam/plumego-4          	  230403	      5340 ns/op	    6561 B/op	      22 allocs/op
BenchmarkRouterMultiParam/chi-4              	  251235	      6351 ns/op	    6084 B/op	      17 allocs/op
BenchmarkRouterMultiParam/chi-4              	  193116	      6379 ns/op	    6084 B/op	      17 allocs/op
BenchmarkRouterMultiParam/chi-4              	  190672	      6636 ns/op	    6084 B/op	      17 allocs/op
BenchmarkRouterMultiParam/gin-4              	  211134	      5973 ns/op	    5380 B/op	      13 allocs/op
BenchmarkRouterMultiParam/gin-4              	  197924	      5779 ns/op	    5380 B/op	      13 allocs/op
BenchmarkRouterMultiParam/gin-4              	  249812	      4740 ns/op	    5380 B/op	      13 allocs/op
BenchmarkRouterMultiParam/echo-4             	  226021	      5504 ns/op	    5379 B/op	      13 allocs/op
BenchmarkRouterMultiParam/echo-4             	  205846	      5778 ns/op	    5379 B/op	      13 allocs/op
BenchmarkRouterMultiParam/echo-4             	  195621	      5256 ns/op	    5379 B/op	      13 allocs/op
BenchmarkRouterScale100/plumego-4            	  221991	      5318 ns/op	    6481 B/op	      21 allocs/op
BenchmarkRouterScale100/plumego-4            	  214425	      5447 ns/op	    6481 B/op	      21 allocs/op
BenchmarkRouterScale100/plumego-4            	  216516	      5162 ns/op	    6481 B/op	      21 allocs/op
BenchmarkRouterScale100/chi-4                	  248156	      4452 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale100/chi-4                	  287208	      4396 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale100/chi-4                	  254640	      4562 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale100/gin-4                	  298174	      4299 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale100/gin-4                	  280086	      4116 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale100/gin-4                	  306752	      4069 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale100/echo-4               	  294026	      4429 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale100/echo-4               	  271267	      4216 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale100/echo-4               	  282306	      4264 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale500/plumego-4            	  226352	      5390 ns/op	    6497 B/op	      21 allocs/op
BenchmarkRouterScale500/plumego-4            	  228730	      5673 ns/op	    6497 B/op	      21 allocs/op
BenchmarkRouterScale500/plumego-4            	  218766	      5608 ns/op	    6497 B/op	      21 allocs/op
BenchmarkRouterScale500/chi-4                	  246859	      4763 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale500/chi-4                	  260365	      4798 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale500/chi-4                	  235508	      4794 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale500/gin-4                	  293031	      4335 ns/op	    5364 B/op	      13 allocs/op
BenchmarkRouterScale500/gin-4                	  279282	      4364 ns/op	    5364 B/op	      13 allocs/op
BenchmarkRouterScale500/gin-4                	  278613	      4322 ns/op	    5364 B/op	      13 allocs/op
BenchmarkRouterScale500/echo-4               	  272104	      4471 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale500/echo-4               	  253062	      4563 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale500/echo-4               	  230415	      4643 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterParallelPlumego-4             	 1000000	      1107 ns/op	    1312 B/op	      12 allocs/op
BenchmarkRouterParallelPlumego-4             	 1000000	      1105 ns/op	    1312 B/op	      12 allocs/op
BenchmarkRouterParallelPlumego-4             	 1000000	      1116 ns/op	    1312 B/op	      12 allocs/op
BenchmarkRouterParallelChi-4                 	 1519263	       914.8 ns/op	     912 B/op	       8 allocs/op
BenchmarkRouterParallelChi-4                 	 1519981	       822.8 ns/op	     912 B/op	       8 allocs/op
BenchmarkRouterParallelChi-4                 	 1543812	       773.3 ns/op	     912 B/op	       8 allocs/op
BenchmarkRouterParallelGin-4                 	 5251036	       219.2 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelGin-4                 	 5120822	       215.8 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelGin-4                 	 5081962	       229.7 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelEcho-4                	 5123973	       232.7 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelEcho-4                	 5558596	       225.9 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelEcho-4                	 5844915	       213.2 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterNotFoundPlumego-4             	  212151	      5632 ns/op	    6223 B/op	      21 allocs/op
BenchmarkRouterNotFoundPlumego-4             	  201661	      6660 ns/op	    6223 B/op	      21 allocs/op
BenchmarkRouterNotFoundPlumego-4             	  178506	      6514 ns/op	    6223 B/op	      21 allocs/op
BenchmarkRouterNotFoundChi-4                 	  213006	      5803 ns/op	    6566 B/op	      22 allocs/op
BenchmarkRouterNotFoundChi-4                 	  213820	      5509 ns/op	    6566 B/op	      22 allocs/op
BenchmarkRouterNotFoundChi-4                 	  218073	      5438 ns/op	    6566 B/op	      22 allocs/op
BenchmarkRouterNotFoundGin-4                 	  237759	      4809 ns/op	    6132 B/op	      17 allocs/op
BenchmarkRouterNotFoundGin-4                 	  261010	      4755 ns/op	    6132 B/op	      17 allocs/op
BenchmarkRouterNotFoundGin-4                 	  246692	      4840 ns/op	    6132 B/op	      17 allocs/op
BenchmarkRouterNotFoundEcho-4                	  192274	      6158 ns/op	    6630 B/op	      25 allocs/op
BenchmarkRouterNotFoundEcho-4                	  191768	      6254 ns/op	    6630 B/op	      25 allocs/op
BenchmarkRouterNotFoundEcho-4                	  181532	      6075 ns/op	    6630 B/op	      25 allocs/op
PASS
ok  	benchmark	185.919s
```
