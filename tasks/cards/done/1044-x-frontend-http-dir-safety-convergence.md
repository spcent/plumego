# Card 1044: x/frontend http.Dir Safety Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/mount.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0743

Goal:
Converge `RegisterFS` and `NewMountFS` behavior for `http.Dir` inputs with the
safer directory-backed mount path.

Scope:
- Detect `http.Dir` inputs and convert them to the same canonicalized
  directory-backed filesystem used by `RegisterFromDir`.
- Apply directory index fail-fast validation to `http.Dir` inputs.
- Preserve lazy behavior for non-directory custom `http.FileSystem`
  implementations.
- Add regression coverage for `RegisterFS(http.Dir(...))` symlink escape,
  relative path stability, and missing index fail-fast behavior.

Non-goals:
- Do not scan arbitrary custom filesystems.
- Do not change `RegisterFS` behavior for embed/custom non-`http.Dir`
  filesystems.
- Do not change public function names.

Files:
- `x/frontend/mount.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Document that `http.Dir` receives directory-backed safety while arbitrary
`RegisterFS` inputs remain caller-owned.

Done Definition:
- `RegisterFS(r, http.Dir(...))` gets symlink escape protection and index
  fail-fast behavior.
- Custom non-`http.Dir` filesystems remain lazy.
- The listed validation commands pass.

Outcome:
- `http.Dir` inputs passed to `RegisterFS`, `NewMountFS`, or `NewHandlerFS` now
  normalize through the same canonical `localDirFS` path used by
  `RegisterFromDir`.
- `http.Dir` mounts now fail fast for missing index files, survive later working
  directory changes, reject symlink escapes, and use directory-backed
  precompressed variant planning.
- Non-`http.Dir` custom filesystems remain lazy and caller-owned.
- Updated `x/frontend` package and module docs to describe the split contract.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
