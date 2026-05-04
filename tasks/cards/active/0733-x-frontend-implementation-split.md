# Card 0733: x/frontend Implementation Split

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/config.go`
- `x/frontend/mount.go`
- `x/frontend/paths.go`
- `x/frontend/compression.go`
Depends On: 0732

Goal:
Reduce `frontend.go` responsibility density without changing behavior.

Scope:
- Move option/config normalization into `config.go`.
- Move mount and registrar code into `mount.go`.
- Move path and local directory filesystem helpers into `paths.go`.
- Move precompressed negotiation helpers into `compression.go`.
- Keep serving behavior unchanged.

Non-goals:
- Do not change public behavior.
- Do not add new abstractions beyond file-level ownership.
- Do not change tests except as required by package compilation.

Files:
- `x/frontend/frontend.go`
- `x/frontend/config.go`
- `x/frontend/mount.go`
- `x/frontend/paths.go`
- `x/frontend/compression.go`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
No docs change required; this is implementation organization only.

Done Definition:
- `frontend.go` focuses on serving behavior.
- Helper files have clear ownership and no behavior drift.
- The listed validation commands pass.

Outcome:

