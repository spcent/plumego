# Card 1425

Milestone: v1-breaking-normalization
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: active
Primary Module: security
Owned Files:
- security/*
Depends On:
- 1421

Goal:
- Remove lenient security fallbacks and make v1 auth, input, header, and abuse
  behavior fail closed by default.

Scope:
- Enumerate legacy token/header/query fallback behavior before editing.
- Remove permissive parsing or silent allow paths that exist only for
  compatibility.
- Keep timing-safe secret comparison.
- Ensure invalid token, invalid signature, invalid policy, and malformed input
  return deterministic failure.
- Add focused negative tests for changed fail-closed paths.

Non-goals:
- Do not log secrets, tokens, signatures, or private keys.
- Do not add external crypto dependencies.
- Do not move tenant or business policy into stable security packages.

Files:
- security/authn/*
- security/jwt/*
- security/input/*
- security/headers/*
- security/abuse/*

Tests:
- go test -timeout 20s ./security/...
- go vet ./security/...
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update security docs or examples for removed lenient fallback behavior.

Done Definition:
- Removed fallback paths have no callers or tests expecting permissive behavior.
- Negative tests assert fail-closed behavior.
- Security tests and dependency checks pass.

Outcome:
- Removed security leniency retained only for compatibility:
  - `headers.Policy.Apply` now validates first and writes no headers when the
    policy is invalid.
  - `headers.CSPBuilder.Build` now returns an empty header value when any
    configured source value is invalid.
  - `abuse.NewLimiter` now returns a disabled fail-closed limiter for invalid
    explicit configuration instead of falling back to defaults.
  - `abuse.Limiter.Allow` now denies empty keys instead of mapping them to the
    shared `unknown` bucket.
  - `input.SanitizeHTML` and `input.SanitizeSQL` legacy aliases were removed;
    callers use the explicit `BestEffortSanitizeHTML` and
    `BestEffortSanitizeSQL` names.
- Updated security docs, deprecation inventory, and stable API snapshots for
  the v1 breaking surface.
- Validation:
  - `go test -timeout 20s ./security/...`
  - `go vet ./security/...`
  - `go run ./internal/checks/dependency-rules`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/agent-workflow`
  - `go build ./...`
