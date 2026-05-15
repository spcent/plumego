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

