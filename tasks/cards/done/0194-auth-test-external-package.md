# Card 0194

Milestone: contract cleanup
Priority: P3
State: done
Primary Module: contract
Owned Files:
- `contract/auth_test.go`
Depends On: —

Goal:
- Convert `auth_test.go` from `package contract_test` to `package contract` so all
  contract test files use a single consistent test package declaration.

Problem:
Every test file in the `contract` package uses the internal test package:

```
package contract  ← 12 files
```

Except one:

```
package contract_test  ← auth_test.go only
```

The external (`_test`) package is the right choice when a test intentionally
validates only the public API surface (blackbox testing). The internal package
(`package contract`) is the right choice when a test needs access to unexported
symbols or when the entire suite uses a consistent internal style.

In this repo the contract test suite uses `package contract` throughout. Having
a single external file creates an inconsistency with no documented justification.
`auth_test.go` does not exercise any behaviour that requires external isolation —
it tests public functions (`WithPrincipal`, `PrincipalFromContext`, etc.) which
are equally testable from the internal package.

Scope:
- Change the package declaration in `auth_test.go` from `package contract_test`
  to `package contract`.
- Remove the `"github.com/spcent/plumego/contract"` import and update all
  `contract.X` references to `X` (unqualified, since the file is now in the
  same package).
- Verify that the test still compiles and passes.

Non-goals:
- No change to test logic or assertions.
- No change to any other test file.

Files:
- `contract/auth_test.go`

Tests:
- `go test -timeout 20s ./contract/...`

Docs Sync: —

Done Definition:
- `auth_test.go` uses `package contract`.
- No file in `contract/` uses `package contract_test`.
- All tests pass.

Outcome:
