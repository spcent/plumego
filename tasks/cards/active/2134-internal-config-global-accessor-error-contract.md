# Card 2134: Internal Config Global Accessor Error Contract
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: active
Primary Module: internal/config
Owned Files:
- internal/config/global.go
- internal/config/global_test.go
Depends On: none

Goal:
Make global config initialization failures observable through an explicit error
path. `GetGlobalConfig` currently auto-initializes global state and, when
`InitDefault` fails, silently returns a new empty manager. That hides the
initialization error from callers and creates two possible manager instances for
the same failed initialization.

Scope:
- Add an error-returning global accessor for callers that need deterministic
  startup failure handling.
- Preserve existing `GetGlobalConfig` compatibility behavior unless a caller is
  explicitly migrated.
- Add tests for successful initialization, failed initialization, reset behavior,
  and compatibility fallback behavior.
- Keep global convenience functions as thin wrappers over the accessor behavior.

Non-goals:
- Do not remove package-level convenience functions in this card.
- Do not change config source precedence, parsing, file watching, or env loading.
- Do not introduce hidden background initialization.

Files:
- `internal/config/global.go`
- `internal/config/global_test.go`

Tests:
- go test -race -timeout 60s ./internal/config
- go test -timeout 20s ./internal/config
- go vet ./internal/config

Docs Sync:
No docs update required unless exported behavior in reference apps changes.

Done Definition:
- Callers have an explicit accessor that returns the global manager and init
  error together.
- Existing compatibility wrappers remain race-free and test-covered.
- Initialization failure no longer has only a silent fallback path.
- The listed validation commands pass.

Outcome:
