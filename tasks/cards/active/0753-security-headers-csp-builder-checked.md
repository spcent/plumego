# Card 0753

Milestone:
Recipe: specs/change-recipes/feature.yaml
Priority: P2
State: active
Primary Module: security
Owned Files:
- security/headers/headers.go
- security/headers/headers_test.go
- docs/modules/security/README.md

Goal:
Add a fail-closed CSP builder path so invalid CSP sources cannot silently weaken production policies.

Scope:
- Add a checked CSP build API that reports invalid directive values.
- Preserve existing `Build()` compatibility behavior.
- Add tests proving checked build rejects unsafe or semicolon-separated values while valid policies still build.
- Sync security docs.

Non-goals:
- Do not change existing `Build()` output or compatibility filtering.
- Do not introduce a full CSP parser.
- Do not change middleware behavior.

Files:
- security/headers/headers.go
- security/headers/headers_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/headers
- go vet ./security/headers
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required for new public checked API.

Done Definition:
- Production callers have a checked CSP builder path.
- Existing `Build()` compatibility behavior remains source-compatible.
- Targeted tests, vet, and dependency checks pass.
