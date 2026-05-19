# Card 1059

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: done
Primary Module: cmd/plumego devserver config edit
Owned Files: cmd/plumego/internal/configmgr/configmgr.go, cmd/plumego/internal/configmgr/configmgr_test.go, cmd/plumego/internal/devserver/config_edit.go, cmd/plumego/internal/devserver/config_edit_test.go
Depends On: 0745

Goal:
Converge dashboard `.env` editing with CLI config parsing and safe serialization.

Scope:
- Reuse or share `.env` parsing rules between config manager and dashboard config edit.
- Preserve comments/blank lines where feasible while validating editable entries.
- Quote or reject values that cannot be safely serialized.
- Add negative tests for spaces, `#`, quotes, invalid keys, and malformed lines.

Non-goals:
- Do not introduce a third-party dotenv dependency.
- Do not create a full `.env` AST if a smaller shared helper is enough.
- Do not change dashboard route shapes.

Files:
- `cmd/plumego/internal/configmgr/configmgr.go`
- `cmd/plumego/internal/configmgr/configmgr_test.go`
- `cmd/plumego/internal/devserver/config_edit.go`
- `cmd/plumego/internal/devserver/config_edit_test.go`

Tests:
- `go test ./internal/configmgr ./internal/devserver`
- `go test ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- Not required unless dashboard config behavior changes materially.

Done Definition:
- Dashboard and CLI parse `.env` entries consistently.
- Dashboard writes valid `.env` values without silent semantic corruption.
- Invalid editable config input fails closed with structured errors.

Outcome:
- Added shared ordered `.env` entry parsing in `configmgr` and kept
  `ParseEnvFile` backed by the same parser.
- Added env key validation and safe value serialization with quoting for spaces,
  comments, quotes, and backslashes.
- Updated dashboard config edit reads to use the shared parser instead of a
  separate ad hoc parser.
- Updated dashboard config edit writes to reject invalid keys and unsupported
  multiline/control-character values before writing.
- Added round-trip and negative tests for parser, serializer, dashboard read,
  dashboard write, invalid keys, and unsupported values.

Validation:
- `go test ./internal/configmgr ./internal/devserver` from `cmd/plumego`
- `go test ./...` from `cmd/plumego`
- `go vet ./...` from `cmd/plumego`
