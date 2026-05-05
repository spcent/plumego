# Card 0759

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: router
Owned Files: router/metadata.go, router/reverse_routing_group_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0758-router-request-trailing-slash-contract

Goal:
Make reverse URL parameter validation strict enough to catch caller mistakes.

Scope:
- Reject duplicate URL param keys.
- Reject unknown URL param keys that are not present in the named route pattern.
- Preserve existing failures for unknown route, missing params, empty params, and
  unpaired key/value lists.
- Add `URL` and `URLMust` regression coverage for the new failure reasons.
- Sync reverse routing docs.

Non-goals:
- Changing `URL` return type.
- Adding exported diagnostic APIs.
- Changing path escaping behavior.

Files:
- router/metadata.go
- router/reverse_routing_group_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- `URL` returns empty string for duplicate or unknown param keys.
- `URLMust` panics with a specific duplicate/unknown param reason.
- Router targeted tests, race tests, and vet pass.

Outcome:
