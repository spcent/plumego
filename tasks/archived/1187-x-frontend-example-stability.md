# Card 1187: x/frontend Example Stability

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/example_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0755

Goal:
Make package examples stable-quality and aligned with recommended usage.

Scope:
- Stop ignoring `http.ListenAndServe` errors in examples.
- Make embedded filesystem examples use `http.FS`/`fs.Sub` guidance instead of
  showing `http.Dir` as the primary embedded example.
- Keep examples concise and compile-only.

Non-goals:
- Do not add runnable embedded fixture assets if not needed.
- Do not change runtime behavior.

Files:
- `x/frontend/example_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go test -race -timeout 60s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Keep examples consistent with README guidance.

Done Definition:
- Examples do not ignore server errors.
- Embedded FS example no longer misleads users toward `http.Dir`.
- The listed validation commands pass.

Outcome:
- Package examples now route server startup through a helper that checks and
  panics on `http.ListenAndServe` errors.
- `RegisterFS` and `NewMountFS` examples now use `http.FS(fstest.MapFS)` as a
  self-contained stand-in for `embed.FS`/`fs.Sub` rather than `http.Dir`.
- README guidance now checks fallback disk registration errors, and module docs
  explicitly keep embedded examples on the `http.FS` path.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go vet ./x/frontend/...`
