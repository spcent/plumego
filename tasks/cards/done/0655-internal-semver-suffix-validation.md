# Card 0655

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/semver
Owned Files: internal/semver/version.go, internal/semver/version_test.go
Depends On: 0654

Goal:
Reject malformed semantic-version prerelease and metadata suffixes instead of accepting empty or invalid identifiers.

Scope:
- Validate prerelease identifiers are non-empty and contain only SemVer identifier characters.
- Validate metadata identifiers are non-empty and contain only SemVer identifier characters.
- Preserve existing support for `v` prefixes and partial versions (`1`, `1.2`) because tests currently document that compatibility.
- Add focused negative tests for empty suffixes, empty dot-separated identifiers, and invalid characters.

Non-goals:
- Do not make core version parsing fully strict SemVer 2.0.0.
- Do not change comparison semantics.
- Do not remove `MustParse`.

Files:
- internal/semver/version.go
- internal/semver/version_test.go

Tests:
- go test -timeout 20s ./internal/semver
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; this is internal validation hardening.

Done Definition:
- Malformed prerelease and metadata suffixes return parse errors.
- Existing valid version, string, and compare tests continue to pass.

Outcome:
Completed. `Parse` now rejects empty prerelease/metadata suffixes, empty dot-separated suffix parts, and unsupported suffix characters while preserving existing partial-version compatibility.

Validation:
- go test -timeout 20s ./internal/semver
- go test -timeout 20s ./internal/...
- go vet ./internal/...
