# Card 0743

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: router
Owned Files: router/dispatch.go, router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0742-router-head-any-body-contract

Goal:
Harden `ServeHTTP` so nil request URLs do not panic on the public
`http.Handler` surface.

Scope:
- Treat a request with nil `URL` as a bad public input and return a stable
  non-panic response.
- Add a regression test for `ServeHTTP` with `req.URL == nil`.
- Sync docs to state the nil URL defensive behavior.
- Mark active queue empty after the final card completes.

Non-goals:
- Supporting malformed URL parsing in router.
- Changing normal net/http server request behavior.
- Adding new public APIs.

Files:
- router/dispatch.go
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Nil request URL produces a deterministic non-panic response.
- Regression coverage passes.
- Router targeted tests, race tests, and vet pass.
- Active queue is empty after completion.

Outcome:
