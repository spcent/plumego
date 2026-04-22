# Card 2109: REST DB Resource Error Contract

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: high
State: done
Primary Module: x/rest
Owned Files:
- x/rest/resource_db.go
- x/rest/resource_db_test.go
- docs/modules/x-rest/README.md
Depends On: none

Goal:
- Make `DBResourceController` error responses follow one stable contract.
- Replace ad hoc hook/repository/client-visible `err.Error()` messages with stable error codes and safe messages.
- Ensure required, not-found, validation, hook, and repository failures have explicit codes that tests can assert.

Scope:
- Audit `DBResourceController` create, update, get, delete, and list error paths.
- Add small local helpers only if they reduce repeated builder chains without creating a second response path.
- Preserve explicit repository and hook injection.
- Add focused tests for hook failures, validation failures, repository failures, and not-found behavior.

Non-goals:
- Do not change the public `ResourceSpec` or repository interfaces unless the symbol-change protocol is explicitly followed.
- Do not introduce business repository ownership or application bootstrap behavior into `x/rest`.
- Do not change successful response envelopes or route registration behavior.

Files:
- `x/rest/resource_db.go`: normalize error construction and remove raw internal messages from client responses.
- `x/rest/resource_db_test.go`: add or update focused controller error-path tests.
- `docs/modules/x-rest/README.md`: document the stable DB resource error-code expectations if behavior changes.

Tests:
- `go test -race -timeout 60s ./x/rest/...`
- `go test -timeout 20s ./x/rest/...`
- `go vet ./x/rest/...`

Docs Sync:
- Required if error codes, public messages, or documented controller behavior change.

Done Definition:
- `rg -n 'Message\(err\.Error\(\)\)' x/rest/resource_db.go` returns no client-facing DB resource paths.
- All `DBResourceController` HTTP error branches use explicit codes.
- Focused tests cover at least one before-hook, after-hook, repository, validation, and not-found failure.
- The three listed validation commands pass.

Outcome:
- Replaced default DB resource hook errors that exposed raw hook messages with stable contract codes and safe public messages.
- Added focused DB resource controller tests for before-hook, after-hook, repository, validation, and not-found failures.
- Documented the stable safe-message behavior in `docs/modules/x-rest/README.md`.
- Validation passed:
  - `go test -race -timeout 60s ./x/rest/...`
  - `go test -timeout 20s ./x/rest/...`
  - `go vet ./x/rest/...`
