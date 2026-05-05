# Card 0757: x/frontend Low Coverage Branches

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/compression_test.go`
- `x/frontend/response_test.go`
- `x/frontend/mount_test.go`
- `docs/extension-evidence/x-frontend.md`
Depends On: 0756

Goal:
Add focused tests for low-coverage but stable-relevant branches.

Scope:
- Cover custom filesystem lazy variant probing paths.
- Cover `tryOpenFile` stat-error behavior through response behavior.
- Cover index serving error behavior.
- Update evidence with the added branch coverage.

Non-goals:
- Do not chase arbitrary coverage percentage.
- Do not change production behavior.

Files:
- `x/frontend/compression_test.go`
- `x/frontend/response_test.go`
- `x/frontend/mount_test.go`
- `docs/extension-evidence/x-frontend.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go test -race -timeout 60s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Update evidence only.

Done Definition:
- Low-coverage stable-relevant branches have regression coverage.
- No production behavior changes are required.
- The listed validation commands pass.
