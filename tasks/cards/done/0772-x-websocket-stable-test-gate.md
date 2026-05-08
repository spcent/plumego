# Card 0772

Milestone: M-003
Recipe: specs/change-recipes/release-readiness.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/modules/x-websocket/README.md`
- `tasks/cards/active/0738-x-websocket-stable-evidence-readiness.md`
Depends On: 0761-0771

Goal:
- Add the final stable-readiness test and gate package for `x/websocket`.

Problem:
Runtime hardening and API cleanup need a final focused gate before any maturity promotion decision. The module needs tests for auth modes, URL-secret rejection, room validation, stop/broadcast races, write deadlines, shutdown edge cases, and protocol fuzz or negative corpus coverage.

Scope:
- Add or consolidate tests for all stable-readiness contracts changed by cards 0761-0771.
- Add fuzz or corpus-style RFC6455 negative tests where practical.
- Run module tests, race tests, vet, build, and boundary/manifest checks.
- Update evidence docs with the final runtime gate results while keeping governance blockers if release evidence is still missing.

Non-goals:
- Do not promote `x/websocket` without real release evidence and owner sign-off.
- Do not widen scope beyond `x/websocket` and required evidence docs.
- Do not update check baselines casually.

Files:
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/modules/x-websocket/README.md`
- `tasks/cards/active/0738-x-websocket-stable-evidence-readiness.md`

Tests:
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required for final gate results and remaining governance blockers.

Done Definition:
- Stable-readiness contracts have direct test coverage.
- Module race tests and vet pass.
- Evidence docs distinguish runtime readiness from release-governance blockers.

Outcome:
- Recorded the final current-head runtime stable-readiness gate in
  `docs/extension-evidence/x-websocket.md` and kept release-governance blockers
  explicit.
- Updated card 0738 to show runtime readiness is complete but maturity evidence
  remains blocked by release refs, release snapshots, and owner sign-off.
- Verified with `go test -race -timeout 60s ./x/websocket/...`, `go vet
  ./x/websocket/...`, `go build ./...`, the four boundary/manifest checks,
  `extension-beta-evidence`, and `extension-maturity`.
