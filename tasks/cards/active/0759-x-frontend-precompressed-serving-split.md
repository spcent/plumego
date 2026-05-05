# Card 0759: x/frontend Precompressed Serving Split

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/compression.go`
- `x/frontend/compression_test.go`
Depends On: 0758

Goal:
Separate directory-backed planned precompressed serving from custom filesystem
lazy probing and avoid duplicate planned variant attempts.

Scope:
- Keep directory-backed planned variant serving before original asset open.
- Keep custom filesystem lazy probing after original asset validation.
- Avoid retrying already-planned directory variants after the original asset has
  been opened.
- Keep response negotiation and headers unchanged.

Non-goals:
- Do not add public API.
- Do not add runtime compression.
- Do not change custom filesystem lazy semantics.

Files:
- `x/frontend/frontend.go`
- `x/frontend/compression.go`
- `x/frontend/compression_test.go`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go test -race -timeout 60s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
No external docs change expected unless behavior text becomes inaccurate.

Done Definition:
- Directory-backed planned variant logic and custom lazy logic are structurally
  separate.
- Known directory variants are not retried through the lazy path.
- The listed validation commands pass.
