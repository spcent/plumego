# Card 0720

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/dispatch.go, router/router_contract_test.go
Depends On: 0719-router-registration-input-validation

Goal:
Make HEAD fallback body suppression behave correctly for io.ReaderFrom and ResponseController unwrapping.

Scope:
- Ensure noBodyWriter.ReadFrom drains the source and returns the drained byte count.
- Expose the wrapped ResponseWriter through Unwrap for net/http compatibility.
- Add focused HEAD fallback tests.

Non-goals:
- Rewriting response wrapping across middleware.
- Adding response body buffering.
- Changing GET dispatch.

Files:
- router/dispatch.go
- router/router_contract_test.go

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required.

Done Definition:
- HEAD fallback suppresses bodies while preserving write/read byte counts.
- Router tests and vet pass.

Outcome:
Updated HEAD fallback body suppression so ReadFrom drains the source and returns the drained byte count, and added Unwrap for net/http ResponseController compatibility.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
