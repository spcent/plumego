# Card 2159: cmd/plumego Config Edit Save Response DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/devserver/config_edit.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`
- `cmd/plumego/DEV_SERVER.md`
Depends On: none

Goal:
Converge the devserver config-edit save response on a local typed DTO while
preserving the existing local dashboard API fields.

Problem:
`handleConfigEditSave` still builds its success payload with `map[string]any`.
That is inconsistent with the typed `ConfigEditResponse` read payload and with
the newer typed dashboard response DTOs.

Scope:
- Replace the config-edit save success map with a local typed response struct.
- Keep field names unchanged: `success`, `path`, `count`, and `restarted`.
- Add a focused handler test that decodes the contract envelope into the typed
  save response.
- Keep the devserver docs aligned with the typed dashboard response policy.

Non-goals:
- Do not change env parsing, normalization, restart behavior, or path
  resolution.
- Do not move devserver code out of `cmd/plumego`.
- Do not add dependencies.

Files:
- `cmd/plumego/internal/devserver/config_edit.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`
- `cmd/plumego/DEV_SERVER.md`

Tests:
- `go test -race -timeout 60s ./internal/devserver/...` from `cmd/plumego`
- `go test -timeout 20s ./internal/devserver/...` from `cmd/plumego`
- `go vet ./internal/devserver/...` from `cmd/plumego`

Docs Sync:
Update `cmd/plumego/DEV_SERVER.md` only to clarify that config-edit dashboard
responses also follow the typed success DTO policy.

Done Definition:
- Config-edit save no longer uses a one-off success map.
- A focused test covers the typed save response through the contract envelope.
- The listed validation commands pass.

Outcome:
