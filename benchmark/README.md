# Plumego Benchmarks

This directory is a separate Go module used for local benchmark evidence. It
keeps comparison-only dependencies, such as chi, out of the main module.

## Prerequisites

- Go 1.24 or newer.
- A source checkout of `github.com/spcent/plumego`.

## Run

From this directory:

```bash
go test -bench=. -benchmem -count=3 ./...
```

The router benchmarks compare Plumego route matching with chi for static,
single-param, and multi-param path shapes. The middleware benchmarks measure
local `func(http.Handler) http.Handler` chains at 1, 3, and 5 layers with no-op
and JSON final handlers.

## Update Results

When updating release evidence, capture the command output verbatim and write it
to `docs/benchmarks/results-v1.1.0.md` with the machine, Go version, command,
and date used for the run.

Chi is the only comparison target here because it is close to Plumego's
standard-library HTTP model and keeps the benchmark surface narrow.
