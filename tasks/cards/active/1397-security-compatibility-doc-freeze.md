# Card 1397

Milestone: M-004
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: security
Owned Files:
- security/input/input.go
- security/abuse/limiter.go
- docs/modules/security/README.md
- specs/deprecation-inventory.yaml
Depends On:
- 1394

Goal:
Freeze stable security compatibility paths while keeping new-code guidance explicit.

Scope:
- Confirm `SanitizeHTML` and `SanitizeSQL` remain compatibility aliases for `BestEffortSanitizeHTML` and `BestEffortSanitizeSQL`.
- Confirm `abuse.NewLimiter` remains lenient compatibility behavior and `NewLimiterWithConfig` remains the strict production startup path.
- Update security docs and inventory wording if current guidance is incomplete.
- Keep negative-path tests for invalid token, signature, header policy, and limiter config behavior intact.

Non-goals:
- Do not remove legacy aliases.
- Do not weaken fail-closed authentication or verification behavior.
- Do not move tenant session or resilience ownership into stable `security`.

Files:
- security/input/input.go
- security/abuse/limiter.go
- docs/modules/security/README.md
- specs/deprecation-inventory.yaml

Tests:
- go test -race -timeout 60s ./security/...
- go test -timeout 20s ./security/...
- go vet ./security/...

Docs Sync:
- Required for security compatibility and strict-vs-lenient constructor guidance.

Done Definition:
- Security compatibility paths are explicitly documented and inventory-backed.
- New-code guidance points to explicit best-effort sanitizer names and strict limiter construction.
- Target checks pass.

Outcome:
