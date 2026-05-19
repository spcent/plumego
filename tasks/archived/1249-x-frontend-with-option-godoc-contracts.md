# Card 1249: x/frontend With Option Godoc Contracts

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/config.go`
Depends On: 0761

Goal:
Make the remaining exported `With*` option contracts self-describing in
pkg.go.dev.

Scope:
- Document header ownership and unsafe value rejection on `WithHeaders`.
- Document precompressed best-effort downgrade and directory/custom filesystem
  behavior on `WithPrecompressed`.
- Document MIME value safety and extension normalization on `WithMIMETypes`.
- Keep behavior unchanged.

Non-goals:
- Do not add public API.
- Do not change option validation.
- Do not promote module status.

Files:
- `x/frontend/config.go`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Go documentation is the docs target for this card.

Done Definition:
- Remaining option godoc exposes stable-relevant constraints.
- No runtime behavior changes occur.
- The listed validation commands pass.

Outcome:
- Added `WithHeaders` godoc for unsafe value rejection and transport-managed
  header ownership.
- Added `WithPrecompressed` godoc for directory-backed planning, custom
  filesystem lazy probing, best-effort variant misses, and identity refusal.
- Added `WithMIMETypes` godoc for extension normalization and unsafe MIME value
  rejection.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
