# Card 0733

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/internal/configmgr
Owned Files: cmd/plumego/internal/configmgr/configmgr.go, cmd/plumego/internal/configmgr/configmgr_test.go, cmd/plumego/internal/checker/checker.go, cmd/plumego/internal/checker/checker_test.go, cmd/plumego/README.md
Depends On: 0732

Goal:
Align `.env` parsing, config validation, and security checks so CLI diagnostics are consistent.

Scope:
- Support common `.env` forms: `export KEY=VALUE`, quoted values, inline comments, and blank/comment lines.
- Stop swallowing parse errors in load/validate/check paths.
- Use one canonical required-secret list for config validation and security checks.
- Add tests for parser compatibility and shared secret requirements.

Non-goals:
- Do not introduce a third-party dotenv dependency.
- Do not add dynamic schema loading.
- Do not change project runtime config semantics beyond diagnostics.

Files:
- `cmd/plumego/internal/configmgr/configmgr.go`
- `cmd/plumego/internal/configmgr/configmgr_test.go`
- `cmd/plumego/internal/checker/checker.go`
- `cmd/plumego/internal/checker/checker_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./internal/configmgr ./internal/checker ./commands`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- Parser behavior is documented by tests.
- Config validation and security check agree on required secrets.
- Env parse errors are visible to command users.

Outcome:
- Added support for `export`, quoted values, and inline comments in `.env` parsing.
- Added shared required secret diagnostics for config validation and security checks.
- Reported env parse errors in load/validate/security paths and documented parser behavior.

Validation:
- `go test ./internal/configmgr ./internal/checker ./commands`
- `go build .`
