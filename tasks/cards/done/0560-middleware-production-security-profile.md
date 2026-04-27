# Card 0560

Milestone:
Recipe: specs/change-recipes/add-middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- docs/modules/middleware/README.md
- docs/modules/security/README.md
- reference/standard-service/README.md
Depends On: 2269

Goal:
Document a production security middleware profile using existing transport-only middleware and security helpers.

Scope:
- Define the recommended explicit stack: request id, recovery, body limit, timeout, security headers, rate limit, auth where needed, access log, and metrics.
- Keep middleware composition explicit in application wiring.
- Clarify which controls belong to `security` helpers versus transport middleware.
- Add reference links rather than new runtime abstractions.

Non-goals:
- Do not add a prebuilt global middleware bundle.
- Do not move tenant policy into stable middleware.
- Do not change default `core` configuration.

Files:
- `docs/modules/middleware/README.md`
- `docs/modules/security/README.md`
- `reference/standard-service/README.md`

Tests:
- `go test -timeout 20s ./middleware/...`
- `go test -timeout 20s ./security/...`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- Required; this is an operational guidance card.

Done Definition:
- Users can copy a production baseline stack while still wiring every middleware explicitly.
- Docs preserve transport-only middleware and stable security boundaries.

Outcome:
Completed. Documented an explicit production security middleware profile in the
middleware primer, clarified the primitive-versus-adapter split in the security
primer, and added reference-app guidance for extending the canonical layout with
app-local middleware wiring.
