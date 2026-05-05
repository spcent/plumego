# Card 0734

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: security
Owned Files:
- security/headers/headers.go
- security/headers/headers_test.go
- docs/modules/security/README.md

Goal:
Make direct security header policy application use the same validation contract as middleware-backed policy application.

Scope:
- Add an error-returning direct application path for `headers.Policy`.
- Keep the existing `Policy.Apply` public API compatible while preventing invalid enum values from being written.
- Add focused tests proving invalid enum values are not emitted by direct policy application.
- Sync security docs with the direct application contract.

Non-goals:
- Do not move middleware behavior into `security/headers`.
- Do not change default or strict policy values.
- Do not add non-stdlib dependencies.

Files:
- security/headers/headers.go
- security/headers/headers_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/headers
- go vet ./security/headers
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update security module docs for direct header policy validation.

Done Definition:
- Direct callers can use an error-returning apply path.
- Existing `Policy.Apply` no longer writes unsupported enum values.
- Tests cover direct invalid policy application.
- Targeted tests, vet, and dependency checks pass.
