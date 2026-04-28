# Card 0658

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/httputil
Owned Files: internal/httputil/html.go, internal/httputil/html_test.go
Depends On: 0657

Goal:
Remove dangerous HTML attributes even when they are separated by tabs, newlines, or carriage returns.

Scope:
- Teach `SanitizeHTML` attribute removal to recognize HTML whitespace before an attribute name.
- Preserve existing behavior for quoted and unquoted attributes.
- Add focused tests for tab/newline-separated event handler attributes.

Non-goals:
- Do not replace the basic sanitizer with a full HTML parser.
- Do not change the documented warning that this helper is not a sole XSS defense.
- Do not change `security/input` in this card.

Files:
- internal/httputil/html.go
- internal/httputil/html_test.go

Tests:
- go test -timeout 20s ./internal/httputil
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; behavior stays within the existing sanitizer contract.

Done Definition:
- Dangerous attributes after spaces, tabs, newlines, and carriage returns are removed.
- Existing sanitizer tests continue to pass.

Outcome:
Completed. Attribute removal now recognizes HTML whitespace before dangerous attributes, including spaces, tabs, newlines, carriage returns, and form feeds.

Validation:
- go test -timeout 20s ./internal/httputil
- go test -timeout 20s ./internal/...
- go vet ./internal/...
