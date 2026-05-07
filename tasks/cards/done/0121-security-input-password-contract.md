# Card 0121

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/input/input.go
- security/input/input_test.go
- security/password/password.go
- security/password/password_strength_test.go
Depends On:
- 0120

Goal:
Clarify security input helper contracts and make password strength checks consistent for stable use.

Scope:
- Add an explicit URL validation helper for externally fetchable URLs that rejects loopback, private, link-local, and metadata-style IP targets.
- Keep existing `ValidateURL` compatibility but document it as syntax/scheme validation, not SSRF protection.
- Make password minimum length count runes consistently.
- Add maximum password length configuration and tests for oversized inputs.

Non-goals:
- Do not add a full HTML sanitizer dependency.
- Do not replace parameterized SQL guidance.
- Do not change password hash format.

Files:
- security/input/input.go
- security/input/input_test.go
- security/password/password.go
- security/password/password_strength_test.go

Tests:
- go test -timeout 20s ./security/input ./security/password
- go vet ./security/input ./security/password

Docs Sync:
- Update public comments for changed helper contracts.

Done Definition:
- SSRF-sensitive callers have a clearly named helper.
- Existing `ValidateURL` behavior remains compatible.
- Password length checks are rune-based and bounded when configured.

Outcome:
- Added `input.ValidatePublicURL` for SSRF-sensitive absolute HTTP(S) fetch targets, rejecting literal localhost, private, loopback, link-local, metadata, multicast, and unspecified IP targets.
- Clarified `ValidateURL` as a URL shape/scheme helper, not an SSRF boundary.
- Made password strength length checks count Unicode code points rather than bytes.
- Added `MaxLength` to `PasswordStrengthConfig` with a default bound of 1024 code points and fail-closed handling for negative bounds.
- Updated security module docs and stable API snapshot.

Validation:
- go test -timeout 20s ./security/input ./security/password
- go vet ./security/input ./security/password
