# Card 0689

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/headers/headers.go
- security/headers/headers_test.go
- security/input/input.go
- security/input/input_test.go
- docs/modules/security/README.md
Depends On: 0688

Goal:
Close remaining header/input construction gaps that allow policy or sanitizer output to be less strict than the package documentation implies.

Scope:
- Make CSP builder methods defensively copy and filter directive values so semicolon/control-character injection cannot create extra directives.
- Drop unsafe CSP report and sandbox values consistently with source-list values.
- Reject URLs containing userinfo credentials in `ValidateURL`.
- Tighten SQL sanitizer keyword stripping to be case-insensitive for mixed-case injection probes.
- Add focused negative tests for CSP directive injection, URL credentials, and mixed-case SQL patterns.

Non-goals:
- Do not add a non-stdlib HTML or SQL sanitizer.
- Do not change middleware transport wiring.
- Do not add trusted-proxy configuration or application bootstrap policy.

Files:
- `security/headers/headers.go`
- `security/headers/headers_test.go`
- `security/input/input.go`
- `security/input/input_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/headers ./security/input`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for CSP builder filtering and URL credential rejection semantics.

Done Definition:
- CSP builder output cannot be expanded by semicolon or control-character values supplied as sources.
- `ValidateURL` rejects URL userinfo credentials.
- `SanitizeSQL` removes mixed-case dangerous keywords used by existing sanitizer policy.
- Targeted security tests and vet pass.

Outcome:
- Routed all CSP source-list and report/sandbox setters through a common directive sanitizer.
- Dropped CSP directive values containing semicolons or unsafe header controls so caller input cannot create extra directives.
- Added nil-safe CSP builder `Build` behavior and preserved explicit flag directives.
- Updated `ValidateURL` to reject embedded userinfo credentials.
- Changed SQL sanitizer keyword removal to case-insensitive matching for the existing dangerous keyword set.
- Synced the security module primer with CSP filtering and URL credential rejection semantics.
- Validation run: `go test -race -timeout 60s ./security/headers ./security/input`; `go test -timeout 20s ./security/...`; `go vet ./security/...`.
