# Card 0416: File API Route Param Source Convergence
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: high
State: done
Primary Module: x/fileapi
Owned Files:
- x/fileapi/handler.go
- x/fileapi/handler_test.go
- docs/modules/x-fileapi/README.md
Depends On: none

Goal:
Make file API handlers read route parameters from Plumego's canonical request
context instead of `http.Request.PathValue`. The Plumego router stores params in
`contract.RequestContext`, while current file API tests seed `PathValue`
directly, so handler tests do not exercise the integration path used by the
repository router.

Scope:
- Replace direct `r.PathValue("id")` reads in `x/fileapi/handler.go` with a
  small local helper based on `contract.RequestContextFromContext`.
- Update handler tests to seed `contract.WithRequestContext` params instead of
  `req.SetPathValue`.
- Preserve existing HTTP status codes, error codes, and response envelopes.
- Keep x/fileapi free of a router dependency.

Non-goals:
- Do not change file storage behavior, URL generation, or upload validation.
- Do not introduce a route-registration helper.
- Do not change stable public APIs.

Tests:
- go test -race -timeout 60s ./x/fileapi/...
- go test -timeout 20s ./x/fileapi/...
- go vet ./x/fileapi/...

Docs Sync:
Update `docs/modules/x-fileapi/README.md` only if it documents manual test setup
or handler parameter behavior.

Done Definition:
- Production file API handlers no longer call `PathValue`.
- File API tests seed the same parameter context shape used by Plumego router.
- The x/fileapi validation commands pass.
- x/fileapi still does not import `router`.

Outcome:
Implemented canonical route-parameter lookup through
`contract.RequestContextFromContext`. File API handler tests now seed
`contract.WithRequestContext` instead of `http.Request.SetPathValue`, and an
`rg` check confirms no `PathValue` or `SetPathValue` usage remains under
`x/fileapi`.

Validation:
- `go test -race -timeout 60s ./x/fileapi/...`
- `go test -timeout 20s ./x/fileapi/...`
- `go vet ./x/fileapi/...`
- `rg -n "SetPathValue|PathValue\\(" x/fileapi -g '*.go'` returned no matches.
