# Card 0739: x/frontend Accept-Encoding Contract

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/compression.go`
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0738

Goal:
Align precompressed response selection with explicit `Accept-Encoding`
semantics, including identity refusal and invalid quality values.

Scope:
- Treat invalid `q` values as invalid tokens instead of clamping out-of-range
  values into the valid range.
- Respect explicit `identity;q=0` when no acceptable precompressed variant is
  available.
- Preserve wildcard and explicit-token precedence for `br` and `gzip`.
- Add regression tests for `identity`, wildcard, invalid `q`, and fallback to
  original assets.

Non-goals:
- Do not add runtime compression.
- Do not negotiate encodings beyond `br`, `gzip`, and identity.
- Do not change cache-control policy.

Files:
- `x/frontend/compression.go`
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Document supported encoding negotiation, invalid q behavior, and identity
refusal.

Done Definition:
- `identity;q=0` does not silently return an unencoded original response.
- Invalid q values are ignored consistently.
- Existing `br`/`gzip` quality ordering remains covered.
- The listed validation commands pass.

Outcome:
- `Accept-Encoding` tokens with invalid or out-of-range `q` values are ignored
  instead of being clamped.
- Requests that reject `identity` now receive `406 Not Acceptable` when no
  accepted precompressed variant is available.
- Existing `br`/`gzip` ordering and wildcard semantics remain covered.
- Documentation now describes the `identity` refusal behavior.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
