# Card 1504

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: log
Owned Files:
- `log/module.yaml`
- `log/logger.go`
Depends On:

Goal:
- Make the `log` manifest enumerate the exported construction types that
  callers actually use.
- Fix the incorrect `glog.Fields` package reference in the public doc comment.

Scope:
- Update the `log` manifest public surface list and the misleading example
  comment in `log/logger.go`.

Non-goals:
- Do not change logger behavior.
- Do not refactor private `gLogger` naming in this card.
- Do not add new public logging APIs.

Files:
- `log/module.yaml`
- `log/logger.go`

<!-- none; metadata and doc-comment card -->

Tests:
- `go test -timeout 20s ./log/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- `docs/modules/log/README.md` only if the manifest additions require README wording updates.

Validation:
- `go test -timeout 20s ./log/...`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/public-entrypoints-sync`

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
<!-- Agent fills this after completion: what changed and why. -->
