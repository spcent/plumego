# Card 0921

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: done
Primary Module: router
Owned Files: router/metadata.go, router/reverse_routing_group_test.go, docs/modules/router/README.md
Depends On: 0733-router-freeze-runtime-policy-contract

Goal:
Keep reverse routing from generating URLs that cannot match their route.

Scope:
- Make `URL` return empty string when required segment or wildcard params are
  present but empty.
- Keep missing param behavior unchanged.
- Add tests for empty segment and wildcard param values.
- Update reverse routing docs.

Non-goals:
- Changing URL escaping rules for non-empty params.
- Supporting optional path segments.
- Changing `URLMust` panic behavior.

Files:
- router/metadata.go
- router/reverse_routing_group_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Empty required route params produce `""` from URL.
- Existing non-empty reverse routing behavior remains unchanged.
- Router tests, race tests, and vet pass.

Outcome:
- Updated reverse routing to return an empty string for empty required segment
  or wildcard params.
- Added tests for empty segment and wildcard param values.
- Updated router docs for unknown, missing, and empty required params.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
