# Card 0119

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
Depends On:

Goal:
Make JWT key persistence cancellation-aware and reduce global manager lock time during token issuance.

Scope:
- Add a context-aware key-store path while preserving compatibility with existing `KeyStore` implementations.
- Avoid holding the manager write lock while generating token IDs and signing access/refresh tokens.
- Keep key material out of public rotation responses.
- Add focused tests for context cancellation and concurrent verification during issuance.

Non-goals:
- Do not add dependencies.
- Do not remove existing exported symbols.
- Do not introduce session lifecycle or revocation behavior.

Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go

Tests:
- go test -timeout 20s ./security/jwt
- go test -race -timeout 60s ./security/jwt
- go vet ./security/jwt

Docs Sync:
- Not required unless public comments change behavior.

Done Definition:
- Context-aware key-store implementations can abort key load, rotation, and active-key writes.
- Token generation copies the active signing key under lock and signs outside the manager lock.
- Existing `KeyStore` implementations continue to compile.
- JWT focused tests pass.

Outcome:
- Added `ContextKeyStore`, `NewJWTManagerContext`, and `RotateKeyContext` so context-aware key stores can abort JWT startup, read, and write paths.
- Kept the existing `KeyStore`, `NewJWTManager`, and `RotateKey` APIs compatible by routing them through background-context wrappers.
- Reduced `GenerateTokenPair` lock scope by copying the active signing key under the manager lock and signing access/refresh tokens after releasing it.
- Added tests proving context-aware stores are used instead of legacy methods for startup and rotation.

Validation:
- go test -timeout 20s ./security/jwt
- go vet ./security/jwt
- go test -race -timeout 60s ./security/jwt
