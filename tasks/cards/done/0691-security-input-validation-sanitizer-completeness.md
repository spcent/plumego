# Card 0691

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/input/input.go
- security/input/input_test.go
- docs/modules/security/README.md
Depends On: 0690

Goal:
Tighten remaining input-safety checks where the current behavior is looser than the package documentation suggests.

Scope:
- Reject email domain labels containing characters outside ASCII letters, digits, and hyphen.
- Reject email domain labels that are too long for DNS label rules.
- Make basic HTML sanitizer remove multiline script blocks.
- Make basic HTML sanitizer remove unquoted event-handler attributes.
- Add focused tests for invalid email labels and sanitizer evasions.

Non-goals:
- Do not replace the basic sanitizer with a non-stdlib HTML sanitizer.
- Do not attempt full RFC email validation.
- Do not add URL allowlist or SSRF policy behavior to stable security.

Files:
- `security/input/input.go`
- `security/input/input_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/input`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for clarified email and sanitizer semantics.

Done Definition:
- Invalid email domain labels are rejected.
- Multiline script tags and unquoted event-handler attributes are removed by `SanitizeHTML`.
- Targeted security tests and vet pass.

Outcome:
- Added DNS-style email domain label validation for allowed characters and maximum label length.
- Updated the basic HTML sanitizer to remove multiline script blocks.
- Updated the basic HTML sanitizer to remove quoted and unquoted inline event-handler attributes.
- Synced the security module primer with email-domain and sanitizer coverage notes.
- Validation run: `go test -race -timeout 60s ./security/input`; `go test -timeout 20s ./security/...`; `go vet ./security/...`.
