# Card 0737: x/frontend Directory Root and Startup Validation

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/mount.go`
- `x/frontend/paths.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0736

Goal:
Make directory-backed mounts deterministic at startup and independent of later
process working-directory changes.

Scope:
- Store directory-backed roots as absolute canonical filesystem paths.
- Validate that the configured index file can be opened during
  `NewMountFromDir` / `RegisterFromDir`.
- Keep `RegisterFS` lazy because arbitrary `http.FileSystem` implementations may
  represent remote or generated backends.
- Add coverage for relative directory mounts across `chdir` and missing index
  fail-fast behavior.

Non-goals:
- Do not change `RegisterFS` construction semantics.
- Do not add hidden filesystem scanning beyond the configured index.
- Do not change symlink escape policy.

Files:
- `x/frontend/mount.go`
- `x/frontend/paths.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Document directory fail-fast behavior and the lazy `RegisterFS` boundary.

Done Definition:
- Relative directory mounts continue serving after a process `chdir`.
- Missing configured index files fail during directory mount construction.
- Symlink escape coverage still passes.
- The listed validation commands pass.

Outcome:
- `NewMountFromDir` and `RegisterFromDir` now resolve directory roots to
  absolute canonical paths before constructing the handler.
- Directory-backed mounts fail fast when the configured index file is missing or
  is a directory.
- `RegisterFS` remains lazy and unchanged for caller-provided filesystems.
- Documentation now states the directory and `RegisterFS` readiness boundary.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
