# Card 0743

Milestone:
Recipe: specs/change-recipes/doc-sync.yaml
Priority: P2
State: active
Primary Module: security
Owned Files:
- security/input/input.go
- security/input/input_test.go
- docs/modules/security/README.md
- docs/stable-api/snapshots/security-head.snapshot

Goal:
Make legacy sanitizer alias status explicit so callers do not over-trust `SanitizeHTML` or `SanitizeSQL`.

Scope:
- Strengthen comments for `SanitizeHTML` and `SanitizeSQL` as legacy compatibility aliases.
- Add tests proving aliases continue to call `BestEffort*`.
- Update security docs and API snapshot wording if snapshot captures comments.

Non-goals:
- Do not remove or rename exported symbols.
- Do not implement a full HTML sanitizer or SQL parser.
- Do not add non-stdlib dependencies.

Files:
- security/input/input.go
- security/input/input_test.go
- docs/modules/security/README.md
- docs/stable-api/snapshots/security-head.snapshot

Tests:
- go test -timeout 20s ./security/input
- go vet ./security/input
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required for public security guidance.

Done Definition:
- Legacy alias guidance is explicit in code and docs.
- Alias behavior remains unchanged and tested.
- Targeted tests, vet, and dependency checks pass.
