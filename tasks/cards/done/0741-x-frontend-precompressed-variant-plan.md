# Card 0741: x/frontend Precompressed Variant Plan

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/compression.go`
- `x/frontend/frontend.go`
- `x/frontend/mount.go`
- `x/frontend/frontend_test.go`
- `x/frontend/frontend_bench_test.go`
- `x/frontend/README.md`
Depends On: 0740

Goal:
Remove avoidable per-request variant probing for directory-backed mounts while
keeping `RegisterFS` behavior lazy and generic.

Scope:
- Build an immutable precompressed variant plan for directory-backed mounts.
- Use the plan to decide whether `Vary: Accept-Encoding` is needed and which
  variants may be attempted.
- Preserve lazy probing for caller-provided `http.FileSystem` values.
- Refresh benchmarks or add assertions that directory-backed serving uses the
  plan.

Non-goals:
- Do not scan arbitrary `RegisterFS` backends.
- Do not introduce file watchers or live reload.
- Do not add runtime compression.

Files:
- `x/frontend/compression.go`
- `x/frontend/frontend.go`
- `x/frontend/mount.go`
- `x/frontend/frontend_test.go`
- `x/frontend/frontend_bench_test.go`
- `x/frontend/README.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go test -bench=Benchmark -run '^$' ./x/frontend`
- `go vet ./x/frontend/...`

Docs Sync:
Document that directory mounts precompute static variant metadata while
`RegisterFS` remains lazy.

Done Definition:
- Directory-backed mounts no longer need extra variant existence probes on every
  uncompressed asset response.
- `RegisterFS` precompressed behavior remains unchanged.
- Benchmarks compile and run.
- The listed validation commands pass.

Outcome:
- Directory-backed handlers now receive an immutable precompressed variant plan
  built at construction time.
- Uncompressed directory-backed asset responses use the plan to decide whether
  `Vary: Accept-Encoding` is needed without probing variants per request.
- `RegisterFS` and `NewMountFS` keep lazy variant probing for caller-provided
  filesystems.
- Documentation now states the directory plan versus `RegisterFS` lazy boundary.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go test -bench=Benchmark -run '^$' ./x/frontend`
  - `go vet ./x/frontend/...`
