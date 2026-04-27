# Card 6205

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/semver
Owned Files: internal/semver/version.go, internal/semver/version_test.go
Depends On: tasks/cards/done/6204-internal-httpx-client-ip-validation.md

Goal:
Align prerelease comparison with semantic-version identifier ordering.

Scope:
- Compare prerelease identifiers segment-by-segment.
- Compare numeric identifiers numerically.
- Treat numeric identifiers as lower precedence than non-numeric identifiers.
- Add focused tests for numeric prerelease ordering and identifier length.

Non-goals:
- Do not change parser acceptance rules.
- Do not change metadata comparison behavior.
- Do not add dependencies.

Files:
- internal/semver/version.go
- internal/semver/version_test.go

Tests:
- go test ./internal/semver
- go test ./internal/...

Docs Sync:
- None; this fixes existing semver behavior.

Done Definition:
- `1.0.0-2` sorts before `1.0.0-10`.
- Standard prerelease precedence examples pass.
- Focused and internal package tests pass.

Outcome:
