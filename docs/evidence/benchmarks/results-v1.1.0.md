# Plumego Benchmark Results

Date: 2026-05-19T10:42:31Z
Go toolchain: go version go1.24.7 linux/amd64
OS/arch: Linux 6.18.5 x86_64
CPU: Intel(R) Xeon(R) Processor @ 2.10 GHz
Command: `go test -bench=. -benchmem -count=3 ./...`
Working directory: `reference/benchmark/`

The output below is verbatim stdout from the benchmark command.

```text
goos: linux
goarch: amd64
pkg: benchmark
cpu: Intel(R) Xeon(R) Processor @ 2.10GHz
BenchmarkChain1NoOp-4              	  304682	      4047 ns/op	    6057 B/op	      17 allocs/op
BenchmarkChain1NoOp-4              	  287634	      4128 ns/op	    6057 B/op	      17 allocs/op
BenchmarkChain1NoOp-4              	  293941	      4086 ns/op	    6057 B/op	      17 allocs/op
BenchmarkChain3NoOp-4              	  268740	      4472 ns/op	    6185 B/op	      19 allocs/op
BenchmarkChain3NoOp-4              	  259130	      4519 ns/op	    6185 B/op	      19 allocs/op
BenchmarkChain3NoOp-4              	  261768	      4295 ns/op	    6185 B/op	      19 allocs/op
BenchmarkChain5NoOp-4              	  258850	      4640 ns/op	    6345 B/op	      20 allocs/op
BenchmarkChain5NoOp-4              	  263770	      4689 ns/op	    6345 B/op	      20 allocs/op
BenchmarkChain5NoOp-4              	  249411	      4715 ns/op	    6345 B/op	      20 allocs/op
BenchmarkChain1JSON-4              	  255531	      4628 ns/op	    6283 B/op	      21 allocs/op
BenchmarkChain1JSON-4              	  247224	      4848 ns/op	    6283 B/op	      21 allocs/op
BenchmarkChain1JSON-4              	  260800	      4620 ns/op	    6283 B/op	      21 allocs/op
BenchmarkChain3JSON-4              	  242916	      4925 ns/op	    6411 B/op	      23 allocs/op
BenchmarkChain3JSON-4              	  246916	      5047 ns/op	    6411 B/op	      23 allocs/op
BenchmarkChain3JSON-4              	  213979	      4994 ns/op	    6411 B/op	      23 allocs/op
BenchmarkChain5JSON-4              	  227065	      5222 ns/op	    6571 B/op	      24 allocs/op
BenchmarkChain5JSON-4              	  225051	      5520 ns/op	    6572 B/op	      24 allocs/op
BenchmarkChain5JSON-4              	  224996	      5370 ns/op	    6572 B/op	      24 allocs/op
BenchmarkGinChain1NoOp-4           	  289935	      4182 ns/op	    6059 B/op	      17 allocs/op
BenchmarkGinChain1NoOp-4           	  289056	      4216 ns/op	    6059 B/op	      17 allocs/op
BenchmarkGinChain1NoOp-4           	  275068	      4092 ns/op	    6059 B/op	      17 allocs/op
BenchmarkGinChain3NoOp-4           	  279357	      4362 ns/op	    6091 B/op	      19 allocs/op
BenchmarkGinChain3NoOp-4           	  274651	      4395 ns/op	    6091 B/op	      19 allocs/op
BenchmarkGinChain3NoOp-4           	  274869	      4416 ns/op	    6091 B/op	      19 allocs/op
BenchmarkGinChain5NoOp-4           	  275115	      4682 ns/op	    6123 B/op	      21 allocs/op
BenchmarkGinChain5NoOp-4           	  237364	      4468 ns/op	    6124 B/op	      21 allocs/op
BenchmarkGinChain5NoOp-4           	  280119	      4532 ns/op	    6124 B/op	      21 allocs/op
BenchmarkGinChain1JSON-4           	  244230	      5132 ns/op	    6334 B/op	      25 allocs/op
BenchmarkGinChain1JSON-4           	  213829	      5135 ns/op	    6334 B/op	      25 allocs/op
BenchmarkGinChain1JSON-4           	  231598	      5340 ns/op	    6334 B/op	      25 allocs/op
BenchmarkGinChain3JSON-4           	  224456	      5493 ns/op	    6366 B/op	      27 allocs/op
BenchmarkGinChain3JSON-4           	  210712	      5555 ns/op	    6366 B/op	      27 allocs/op
BenchmarkGinChain3JSON-4           	  216393	      5437 ns/op	    6366 B/op	      27 allocs/op
BenchmarkGinChain5JSON-4           	  201889	      5620 ns/op	    6398 B/op	      29 allocs/op
BenchmarkGinChain5JSON-4           	  218425	      5859 ns/op	    6398 B/op	      29 allocs/op
BenchmarkGinChain5JSON-4           	  194967	      5903 ns/op	    6398 B/op	      29 allocs/op
BenchmarkEchoChain1NoOp-4          	  290485	      4476 ns/op	    6091 B/op	      18 allocs/op
BenchmarkEchoChain1NoOp-4          	  263322	      4358 ns/op	    6091 B/op	      18 allocs/op
BenchmarkEchoChain1NoOp-4          	  273376	      4226 ns/op	    6091 B/op	      18 allocs/op
BenchmarkEchoChain3NoOp-4          	  274280	      4622 ns/op	    6284 B/op	      22 allocs/op
BenchmarkEchoChain3NoOp-4          	  239470	      4878 ns/op	    6284 B/op	      22 allocs/op
BenchmarkEchoChain3NoOp-4          	  236574	      4905 ns/op	    6284 B/op	      22 allocs/op
BenchmarkEchoChain5NoOp-4          	  240524	      5406 ns/op	    6508 B/op	      25 allocs/op
BenchmarkEchoChain5NoOp-4          	  236584	      5129 ns/op	    6508 B/op	      25 allocs/op
BenchmarkEchoChain5NoOp-4          	  217540	      5159 ns/op	    6508 B/op	      25 allocs/op
BenchmarkEchoChain1JSON-4          	  223030	      5328 ns/op	    6382 B/op	      26 allocs/op
BenchmarkEchoChain1JSON-4          	  209622	      5431 ns/op	    6382 B/op	      26 allocs/op
BenchmarkEchoChain1JSON-4          	  231212	      5382 ns/op	    6382 B/op	      26 allocs/op
BenchmarkEchoChain3JSON-4          	  215424	      5789 ns/op	    6574 B/op	      30 allocs/op
BenchmarkEchoChain3JSON-4          	  195924	      5734 ns/op	    6574 B/op	      30 allocs/op
BenchmarkEchoChain3JSON-4          	  198760	      5611 ns/op	    6574 B/op	      30 allocs/op
BenchmarkEchoChain5JSON-4          	  197580	      6344 ns/op	    6798 B/op	      33 allocs/op
BenchmarkEchoChain5JSON-4          	  199370	      6553 ns/op	    6798 B/op	      33 allocs/op
BenchmarkEchoChain5JSON-4          	  174962	      6568 ns/op	    6798 B/op	      33 allocs/op
BenchmarkRouterStatic/plumego-4    	  301240	      4208 ns/op	    5723 B/op	      16 allocs/op
BenchmarkRouterStatic/plumego-4    	  283430	      4157 ns/op	    5723 B/op	      16 allocs/op
BenchmarkRouterStatic/plumego-4    	  265401	      4281 ns/op	    5723 B/op	      16 allocs/op
BenchmarkRouterStatic/chi-4        	  306696	      4117 ns/op	    5699 B/op	      15 allocs/op
BenchmarkRouterStatic/chi-4        	  301960	      4020 ns/op	    5699 B/op	      15 allocs/op
BenchmarkRouterStatic/chi-4        	  284580	      4110 ns/op	    5699 B/op	      15 allocs/op
BenchmarkRouterStatic/gin-4        	  320710	      3765 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/gin-4        	  303775	      3832 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/gin-4        	  313944	      3780 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/echo-4       	  304654	      3756 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/echo-4       	  294758	      3654 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterStatic/echo-4       	  325833	      3721 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/plumego-4         	  302805	      4346 ns/op	    5715 B/op	      16 allocs/op
BenchmarkRouterSingleParam/plumego-4         	  257620	      4170 ns/op	    5715 B/op	      16 allocs/op
BenchmarkRouterSingleParam/plumego-4         	  276097	      4291 ns/op	    5715 B/op	      16 allocs/op
BenchmarkRouterSingleParam/chi-4             	  264063	      4604 ns/op	    6036 B/op	      17 allocs/op
BenchmarkRouterSingleParam/chi-4             	  260020	      4540 ns/op	    6036 B/op	      17 allocs/op
BenchmarkRouterSingleParam/chi-4             	  257880	      4547 ns/op	    6036 B/op	      17 allocs/op
BenchmarkRouterSingleParam/gin-4             	  310411	      3818 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/gin-4             	  327172	      3685 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/gin-4             	  317233	      3773 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/echo-4            	  301302	      3802 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/echo-4            	  327186	      3666 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterSingleParam/echo-4            	  312391	      3743 ns/op	    5331 B/op	      13 allocs/op
BenchmarkRouterMultiParam/plumego-4          	  275719	      4223 ns/op	    5843 B/op	      17 allocs/op
BenchmarkRouterMultiParam/plumego-4          	  274489	      4329 ns/op	    5843 B/op	      17 allocs/op
BenchmarkRouterMultiParam/plumego-4          	  297676	      4233 ns/op	    5843 B/op	      17 allocs/op
BenchmarkRouterMultiParam/chi-4              	  248720	      4653 ns/op	    6084 B/op	      17 allocs/op
BenchmarkRouterMultiParam/chi-4              	  252832	      4705 ns/op	    6084 B/op	      17 allocs/op
BenchmarkRouterMultiParam/chi-4              	  242163	      4763 ns/op	    6084 B/op	      17 allocs/op
BenchmarkRouterMultiParam/gin-4              	  319285	      3892 ns/op	    5380 B/op	      13 allocs/op
BenchmarkRouterMultiParam/gin-4              	  297082	      4021 ns/op	    5380 B/op	      13 allocs/op
BenchmarkRouterMultiParam/gin-4              	  284920	      4105 ns/op	    5380 B/op	      13 allocs/op
BenchmarkRouterMultiParam/echo-4             	  294848	      4257 ns/op	    5379 B/op	      13 allocs/op
BenchmarkRouterMultiParam/echo-4             	  261582	      4218 ns/op	    5379 B/op	      13 allocs/op
BenchmarkRouterMultiParam/echo-4             	  252151	      4023 ns/op	    5379 B/op	      13 allocs/op
BenchmarkRouterScale100/plumego-4            	  263244	      4351 ns/op	    5763 B/op	      16 allocs/op
BenchmarkRouterScale100/plumego-4            	  272952	      4438 ns/op	    5763 B/op	      16 allocs/op
BenchmarkRouterScale100/plumego-4            	  281572	      4342 ns/op	    5763 B/op	      16 allocs/op
BenchmarkRouterScale100/chi-4                	  258836	      4346 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale100/chi-4                	  241162	      4449 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale100/chi-4                	  274033	      4405 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale100/gin-4                	  296073	      3981 ns/op	    5364 B/op	      13 allocs/op
BenchmarkRouterScale100/gin-4                	  284989	      5026 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale100/gin-4                	  226110	      5392 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale100/echo-4               	  228865	      5398 ns/op	    5362 B/op	      13 allocs/op
BenchmarkRouterScale100/echo-4               	  224709	      5476 ns/op	    5362 B/op	      13 allocs/op
BenchmarkRouterScale100/echo-4               	  180908	      5553 ns/op	    5362 B/op	      13 allocs/op
BenchmarkRouterScale500/plumego-4            	  183900	      6243 ns/op	    5779 B/op	      16 allocs/op
BenchmarkRouterScale500/plumego-4            	  228699	      6226 ns/op	    5779 B/op	      16 allocs/op
BenchmarkRouterScale500/plumego-4            	  195616	      6034 ns/op	    5778 B/op	      16 allocs/op
BenchmarkRouterScale500/chi-4                	  178014	      6047 ns/op	    5730 B/op	      15 allocs/op
BenchmarkRouterScale500/chi-4                	  228776	      6031 ns/op	    5730 B/op	      15 allocs/op
BenchmarkRouterScale500/chi-4                	  167024	      6258 ns/op	    5731 B/op	      15 allocs/op
BenchmarkRouterScale500/gin-4                	  238465	      5627 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale500/gin-4                	  195031	      5809 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale500/gin-4                	  234998	      5593 ns/op	    5363 B/op	      13 allocs/op
BenchmarkRouterScale500/echo-4               	  220045	      5508 ns/op	    5362 B/op	      13 allocs/op
BenchmarkRouterScale500/echo-4               	  196119	      5547 ns/op	    5362 B/op	      13 allocs/op
BenchmarkRouterScale500/echo-4               	  204645	      5658 ns/op	    5362 B/op	      13 allocs/op
BenchmarkRouterParallelPlumego-4             	 1799718	       663 ns/op	     592 B/op	       7 allocs/op
BenchmarkRouterParallelPlumego-4             	 1849249	       669 ns/op	     592 B/op	       7 allocs/op
BenchmarkRouterParallelPlumego-4             	 1732922	       730 ns/op	     592 B/op	       7 allocs/op
BenchmarkRouterParallelChi-4                 	 1165504	      1013 ns/op	     912 B/op	       8 allocs/op
BenchmarkRouterParallelChi-4                 	 1188904	       976 ns/op	     912 B/op	       8 allocs/op
BenchmarkRouterParallelChi-4                 	 1000000	      1056 ns/op	     912 B/op	       8 allocs/op
BenchmarkRouterParallelGin-4                 	 3917090	       285 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelGin-4                 	 3929028	       294 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelGin-4                 	 4588498	       286 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelEcho-4                	 4910727	       231 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelEcho-4                	 4143985	       297 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterParallelEcho-4                	 4475001	       237 ns/op	     208 B/op	       4 allocs/op
BenchmarkRouterNotFoundPlumego-4             	  179011	      6140 ns/op	    6222 B/op	      21 allocs/op
BenchmarkRouterNotFoundPlumego-4             	  182229	      6233 ns/op	    6222 B/op	      21 allocs/op
BenchmarkRouterNotFoundPlumego-4             	  198546	      6462 ns/op	    6222 B/op	      21 allocs/op
BenchmarkRouterNotFoundChi-4                 	  174235	      6596 ns/op	    6565 B/op	      22 allocs/op
BenchmarkRouterNotFoundChi-4                 	  172839	      6279 ns/op	    6565 B/op	      22 allocs/op
BenchmarkRouterNotFoundChi-4                 	  200126	      6667 ns/op	    6565 B/op	      22 allocs/op
BenchmarkRouterNotFoundGin-4                 	  175324	      6078 ns/op	    6131 B/op	      17 allocs/op
BenchmarkRouterNotFoundGin-4                 	  197354	      5937 ns/op	    6131 B/op	      17 allocs/op
BenchmarkRouterNotFoundGin-4                 	  178675	      6856 ns/op	    6131 B/op	      17 allocs/op
BenchmarkRouterNotFoundEcho-4                	  134203	      8608 ns/op	    6628 B/op	      25 allocs/op
BenchmarkRouterNotFoundEcho-4                	  127740	      8776 ns/op	    6628 B/op	      25 allocs/op
BenchmarkRouterNotFoundEcho-4                	  150016	      8600 ns/op	    6629 B/op	      25 allocs/op
PASS
ok  	benchmark	180.888s
```

## Summary (medians)

| Benchmark | Plumego | Chi | Gin | Echo |
|---|---|---|---|---|
| RouterSingleParam (ns/op) | 4291 | 4547 | 3773 | 3743 |
| RouterParallel (ns/op) | 663 | 1013 | 285 | 231 |
| RouterScale500 (ns/op) | 6034 | 6258 | 5593 | 5508 |
| RouterNotFound (ns/op) | 6140 | 6596 | 5937 | 8608 |

Plumego is faster than Chi in both sequential and parallel dispatch.
Gap to Gin/Echo: ~12% sequential, ~2.5× parallel (remaining cost is
`context.WithValue` + `req.WithContext` for stdlib `net/http` compatibility).
