# Card 0715

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: security
Owned Files:
- security/headers/headers.go
- security/headers/headers_test.go
- security/input/input.go
- docs/modules/security/README.md
- docs/stable-api/snapshots/security-head.snapshot
Depends On: 0714

Goal:
Make security header policy misconfiguration inspectable before runtime and reduce over-broad input sanitizer documentation claims.

Scope:
- Add a `Policy.Validate` method that reports invalid configured header names and values.
- Keep `Policy.Apply` behavior backward-compatible by continuing to skip unsafe runtime header values.
- Add focused tests for policy validation.
- Clarify `security/input` package comments so basic sanitizers are not described as comprehensive RFC or sanitizer implementations.
- Update security docs and the stable API snapshot.

Non-goals:
- Do not change middleware behavior.
- Do not reject responses at runtime from `Policy.Apply`.
- Do not introduce a third-party sanitizer.

Files:
- `security/headers/headers.go`
- `security/headers/headers_test.go`
- `security/input/input.go`
- `docs/modules/security/README.md`
- `docs/stable-api/snapshots/security-head.snapshot`

Tests:
- `go test -race -timeout 60s ./security/headers ./security/input`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for policy validation and input helper semantics.

Done Definition:
- Invalid header policy configuration can be detected with `Policy.Validate`.
- Existing `Apply` fail-closed skip behavior remains intact.
- Input package documentation no longer over-promises comprehensive sanitizer or RFC compliance.
- Targeted tests, vet, and the security API snapshot are updated.

Outcome:

