# Card 6201

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: internal/httputil
Owned Files: internal/httputil/html.go, internal/httputil/html_test.go
Depends On:

Goal:
Close the case-sensitive protocol-removal gap in the basic HTML sanitizer.

Scope:
- Remove `javascript:` and `data:` protocol markers case-insensitively.
- Preserve existing tag and event-handler behavior.
- Add focused tests for mixed-case protocol payloads.

Non-goals:
- Do not replace the sanitizer with a non-stdlib dependency.
- Do not claim full HTML sanitization guarantees.
- Do not change escaping helpers.

Files:
- internal/httputil/html.go
- internal/httputil/html_test.go

Tests:
- go test ./internal/httputil
- go test ./internal/...

Docs Sync:
- None; this tightens existing documented behavior.

Done Definition:
- Mixed-case dangerous protocols are removed.
- Existing sanitizer tests continue to pass.

Outcome:
- Added case-insensitive dangerous protocol removal for the basic sanitizer.
- Added mixed-case `javascript:` and `data:` regression coverage.
- Validation: `go test ./internal/httputil`; `go test ./internal/...`.
