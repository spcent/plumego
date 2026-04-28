# Card 0668

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: internal/config
Owned Files: internal/config/global.go, internal/config/env_test.go
Depends On: 0667

Goal:
Allow package-level `.env` loading to handle valid long environment values.

Scope:
- Configure the `LoadEnvFile` scanner with an explicit larger token limit.
- Add focused coverage for a long single-line `.env` value.
- Preserve overwrite behavior and parse rules.

Non-goals:
- Do not change `FileSource` parsing, which already reads the full file content.
- Do not redesign `.env` syntax handling.

Files:
- internal/config/global.go
- internal/config/env_test.go

Tests:
- go test -timeout 20s ./internal/config
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; internal robustness only.

Done Definition:
- `LoadEnvFile` accepts a valid environment line longer than the scanner default token limit.
- Existing overwrite and error propagation tests still pass.

Outcome:

