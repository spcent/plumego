# Card 0745

Milestone:
Recipe: specs/change-recipes/doc-sync.yaml
Priority: P2
State: active
Primary Module: security
Owned Files:
- security/headers/headers.go
- security/headers/headers_test.go
- docs/modules/security/README.md

Goal:
Prefer fail-closed header policy application in examples and docs.

Scope:
- Update production-facing examples to use middleware or `ApplyChecked`.
- Keep compatibility examples for `Apply` clearly labeled as lenient.
- Add/adjust tests if examples are compiled or asserted.

Non-goals:
- Do not change header application behavior.
- Do not modify middleware packages.
- Do not change default policies.

Files:
- security/headers/headers.go
- security/headers/headers_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/headers
- go vet ./security/headers
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required for security guidance.

Done Definition:
- Examples do not imply lenient `Apply` is the production fail-closed path.
- Targeted tests, vet, and dependency checks pass.
