# Card 2149: security/input Regex Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: security
Owned Files:
- `security/input/input.go`
- `security/input/input_test.go`
- `docs/modules/security/README.md`
Depends On: none

Goal:
Remove repeated regex compilation from `security/input` sanitizers while
preserving the current public behavior.

Problem:
`security/input/input.go` recompiles several fixed regular expressions on every
call to `SanitizeHTML`, `SanitizeSQL`, and whitespace normalization helpers.
The patterns are static, security-sensitive primitives in a stable package, so
they should have one obvious package-local definition and be covered by
negative-path tests. This is not a correctness bug today, but the current shape
is inefficient and makes future sanitizer changes easier to miss in one path.

Scope:
- Move fixed sanitizer regexes to package-local precompiled variables.
- Keep behavior byte-for-byte compatible for existing sanitizer tests.
- Add focused tests for HTML script/event/data/javascript stripping and SQL
  comment stripping if current tests do not already cover those paths.
- Keep dynamic caller-supplied regex behavior out of this card.

Non-goals:
- Do not introduce external sanitizer dependencies.
- Do not expand sanitizer scope or claim production-grade HTML sanitization.
- Do not change stable public APIs.
- Do not touch `internal/validator` in this card.

Files:
- `security/input/input.go`
- `security/input/input_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/input`
- `go test -timeout 20s ./security/input`
- `go vet ./security/input`

Docs Sync:
Update `docs/modules/security/README.md` only if sanitizer behavior or caveats
change; otherwise no docs change is required.

Done Definition:
- Fixed sanitizer regexes are compiled once at package scope.
- Existing sanitizer behavior remains compatible.
- Focused sanitizer tests cover the moved regex paths.
- The listed validation commands pass.

Outcome:
