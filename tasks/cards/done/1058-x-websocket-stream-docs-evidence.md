# Card 1058

Milestone:
Recipe: specs/change-recipes/doc-sync.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/stream.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`
Depends On: 0744, 0724, 0738

Goal:
- Resolve the remaining stream API, maturity wording, and stable evidence blockers before any stable promotion decision.

Problem:
`ReadMessageStream` still sounds like true streaming even though it buffers fragmented payloads into memory under a bound. Package comments also mix experimental status with "production-ready" wording. Evidence remains incomplete because release refs, API snapshots, and owner sign-off are missing.

Scope:
- Rename or replace `ReadMessageStream` so the public name reflects bounded buffered reading, or implement a true streaming reader with clear backpressure and size semantics.
- Delete the old misleading exported stream symbol after updating all callers.
- Remove "production-ready" wording while `x/websocket` remains experimental.
- Update docs to state the exact large-message behavior, memory bounds, and stable-readiness caveats.
- Update `docs/extension-evidence/x-websocket.md` with remaining blockers only; do not fabricate release refs or sign-off.
- Prepare the final evidence checklist for stable promotion after real release artifacts exist.

Non-goals:
- Do not promote `x/websocket` to beta or stable without real release evidence and owner sign-off.
- Do not create synthetic release snapshots.
- Do not add non-standard streaming dependencies.

Files:
- `x/websocket/stream.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
- Required for stream API behavior, maturity wording, evidence blockers, and promotion checklist.

Done Definition:
- The stream API name and implementation no longer imply unbounded or true streaming behavior incorrectly.
- Old misleading exported stream symbols have no stale production call sites.
- Experimental/stable wording is consistent across package comments, module docs, and evidence docs.
- Evidence docs clearly separate completed runtime hardening from external release/sign-off blockers.

Outcome:
- Renamed `ReadMessageStream` to `ReadMessageReader` and removed the old exported symbol.
- Updated server read-loop logs, tests, and complete-message reads to use the bounded reader name.
- Removed `production-ready` wording from experimental package comments.
- Updated module docs and evidence docs to describe bounded read semantics and remaining release-governance blockers.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go run ./internal/checks/extension-beta-evidence`
  - `go build ./...`
