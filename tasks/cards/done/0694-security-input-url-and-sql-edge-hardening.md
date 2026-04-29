# Card 0694

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/input/input.go
- security/input/input_test.go
- docs/modules/security/README.md
Depends On: 0693

Goal:
Close remaining input-safety edge cases where helper behavior is looser than the security-focused package contract.

Scope:
- Make `ValidateURL` apply control-character and parsing checks to root-relative URLs, not only absolute URLs.
- Reject root-relative URLs with raw backslashes or ambiguous network-path form.
- Make `SanitizeSQL` remove multiline block comments.
- Add focused regression tests for relative URL controls/backslashes and multiline SQL comments.

Non-goals:
- Do not add SSRF, CIDR, or application allowlist policy to stable `security/input`.
- Do not replace the basic SQL sanitizer with a SQL parser.
- Do not change exported symbol names.

Files:
- `security/input/input.go`
- `security/input/input_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/input`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for clarified `ValidateURL` and `SanitizeSQL` behavior.

Done Definition:
- Relative URLs containing control characters, raw backslashes, or network-path form are rejected.
- Multiline SQL block comments are removed by `SanitizeSQL`.
- Targeted security tests and vet pass.

Outcome:
- Added relative URL safety checks for network-path URLs, raw control bytes, raw backslashes, and parsed relative-path shape.
- Updated SQL block-comment sanitization to cover multiline comments.
- Added regression tests for unsafe relative URL forms and multiline SQL block comments.
- Synced the security module primer with clarified URL and SQL sanitizer behavior.
- Validation run: `go test -race -timeout 60s ./security/input`; `go test -timeout 20s ./security/...`; `go vet ./security/...`.
