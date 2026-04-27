# Card 4105: Log Second-Pass Documentation Sync

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: done
Primary Module: log
Owned Files:
- `docs/modules/log/README.md`
- `log/module.yaml`
- `tasks/cards/active/README.md`
Depends On:
- `tasks/cards/done/4101-log-writer-snapshot-close-order.md`
- `tasks/cards/done/4102-log-text-field-key-escaping.md`
- `tasks/cards/done/4103-log-json-nested-field-normalization.md`
- `tasks/cards/done/4104-log-json-derived-error-reporting.md`

Goal:
Align log docs and the active queue after the second cleanup pass.

Problem:
The second pass tightened lower-level backend and field-normalization behavior.
The module docs and review guidance should describe only the implemented stable
contract and keep internal backend details out of application-facing guidance.

Scope:
- Update docs for text key safety and recursive JSON fallback if needed.
- Keep module manifest guidance within manifest size limits.
- Move the final second-pass log card to done.
- Run boundary and manifest checks.

Non-goals:
- Do not document unimplemented APIs.
- Do not promote internal file backend helpers to stable API.
- Do not change unrelated non-log task cards.

Outcome:
- Confirmed `docs/modules/log/README.md` already describes the second-pass text
  key and nested JSON fallback behavior.
- Tightened `log/module.yaml` review guidance to include nested JSON field
  preservation while staying within manifest list limits.
- Removed the final log cleanup card from the active queue.
- Left unrelated active store cards untouched.

Validation:
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`

Done Definition:
- Documentation matches the second-pass implementation.
- Active queue no longer lists second-pass log cards.
- Boundary and manifest checks pass.
