# Card 0728: x/frontend API and Response Policy Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
- `x/frontend/embedded_fs.go`
Depends On: 0727

Goal:
Clarify frontend constructor behavior and separate asset response policy from
custom error page policy.

Scope:
- Avoid applying options twice during mount construction.
- Validate configured custom error and not-found page paths with the same
  conservative file-path rules used for served assets.
- Prevent custom 404/5xx pages from inheriting immutable asset cache policy.
- Reassess the package-owned embedded helper surface and keep behavior explicit.

Non-goals:
- Do not remove exported symbols without following the symbol-change protocol.
- Do not add application bootstrap behavior.
- Do not add dependencies.

Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
- `x/frontend/embedded_fs.go`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
May require docs updates if custom page policy or embedded helper guidance
changes.

Done Definition:
- Mount construction derives one canonical config per call.
- Custom page paths are validated before serving.
- Error pages do not receive long asset cache headers by default.
- The listed validation commands pass.

Outcome:
- Refactored mount construction so options are applied once per constructor
  call.
- Added validation for custom not-found and error page paths.
- Split file response policy so custom 404/5xx pages do not inherit asset cache
  headers.
- Kept the existing embedded helper surface unchanged for docs clarification in
  the follow-up docs card.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
