# Card 1075

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/metadata.go, router/reverse_routing_group_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0746-router-static-symlink-contract

Goal:
Make reverse URL generation failure reasons explicit enough for stable
debugging.

Scope:
- Introduce an unexported reverse URL helper that returns a failure reason.
- Keep public `URL` behavior returning empty string on failure.
- Make `URLMust` panic messages distinguish unknown route, missing params,
  empty params, and malformed key/value params.
- Add regression tests for the differentiated panic messages.
- Sync reverse routing docs for `URLMust`.

Non-goals:
- Adding new public APIs.
- Changing successful URL output.
- Changing `URL` empty-string compatibility.

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
- `URL` remains compatible.
- `URLMust` gives actionable panic messages for distinct failure classes.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Added an unexported reverse URL helper that returns both the generated URL
  and a specific failure reason.
- Kept public `URL` failure behavior as an empty string.
- Updated `URLMust` to panic with distinct reasons for unknown route, missing
  required params, empty required params, and unpaired key/value params.
- Added regression coverage for malformed key/value params and `URLMust`
  failure messages.
- Documented the reverse-routing failure contract.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
