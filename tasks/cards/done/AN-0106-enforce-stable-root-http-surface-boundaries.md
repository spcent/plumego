# Card AN-0106

Priority: P0

Goal:
- Add a mechanical check that fails when stable roots grow app-facing HTTP surface.

Scope:
- stable-root package scanning for app-facing HTTP handler leakage
- route registration leakage from stable roots into extension-like API surface
- actionable failure messaging for likely app-facing HTTP surface violations

Non-goals:
- Do not encode every possible HTTP helper as a violation in this card.
- Do not enforce protocol-family namespace rules in this card.
- Do not redesign existing stable package layouts beyond what is needed to support the new check.

Files:
- `internal/checks/agent-workflow/main.go`
- `internal/checks/checkutil/checkutil.go`
- `internal/checks/checkutil/checkutil_test.go`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Keep the new check aligned with the stable-root boundary rules in `AGENTS.md`, `docs/CANONICAL_STYLE_GUIDE.md`, and `docs/ROADMAP_AGENT_NATIVE.md`.

Done Definition:
- Stable-root packages fail checks when they expose clearly app-facing HTTP handlers or route registration helpers.
- The check stays narrow enough to avoid flagging ordinary transport primitives in stable middleware and contract helpers.
- Violations point reviewers and agents to the leaking package path instead of relying on architecture rediscovery during review.

Outcome:
- `internal/checks/checkutil/checkutil.go` now scans non-`core`/`router` stable roots for suspicious handler files and route registration helper symbols.
- `internal/checks/agent-workflow/main.go` now fails when stable roots expose likely app-facing HTTP surface.
- `internal/checks/checkutil/checkutil_test.go` now covers handler leakage detection, route registration leakage detection, and false-positive avoidance for `core` and `metrics`.

Validation Run:
- `go test ./internal/checks/...`
- `go run ./internal/checks/agent-workflow`
