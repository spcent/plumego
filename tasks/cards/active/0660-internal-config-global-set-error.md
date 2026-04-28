# Card 0660

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/config
Owned Files: internal/config/global.go, internal/config/global_test.go
Depends On:

Goal:
Make global config `Set` return environment mutation errors instead of silently continuing.

Scope:
- Return `os.Setenv` errors from `Set`.
- Add a focused test using an invalid environment key.

Non-goals:
- Do not change global initialization compatibility behavior.
- Do not alter reload semantics after a successful set.

Files:
- internal/config/global.go
- internal/config/global_test.go

Tests:
- go test -timeout 20s ./internal/config
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; this makes an existing error observable.

Done Definition:
- `Set` returns immediately when `os.Setenv` fails.
- Config tests and internal validation pass.

Outcome:
