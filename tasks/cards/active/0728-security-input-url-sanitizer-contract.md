# Card 0728

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: security
Owned Files:
- security/input/input.go
- security/input/input_test.go
- docs/modules/security/README.md
Depends On:
- 0727

Goal:
Tighten input URL port validation and make best-effort sanitizer semantics explicit.

Scope:
- Reject non-numeric URL ports in `ValidateURL`.
- Add clearer best-effort sanitizer aliases or comments for HTML/SQL cleanup helpers.
- Add focused negative tests for service-name ports and sanitizer alias behavior.

Non-goals:
- Do not add a full HTML sanitizer.
- Do not perform DNS resolution in `ValidatePublicURL`.
- Do not remove existing exported sanitizer names in this change.

Files:
- security/input/input.go
- security/input/input_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/input
- go vet ./security/input

Docs Sync:
- Clarify that sanitizer helpers are lossy best-effort utilities and not security boundaries.

Done Definition:
- URL ports are strict decimal ports.
- Existing sanitizer APIs remain compatible.
- Documentation names the safe usage boundary.

Outcome:

Validation:
