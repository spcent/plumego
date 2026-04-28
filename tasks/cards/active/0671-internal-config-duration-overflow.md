# Card 0671

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/config
Owned Files: internal/config/manager.go, internal/config/config_test.go
Depends On:

Goal:
Prevent struct unmarshal duration fallback from silently overflowing when numeric values are interpreted as milliseconds.

Scope:
- Check numeric millisecond values before multiplying by `time.Millisecond`.
- Return a clear error when the duration would overflow `time.Duration`.
- Add focused regression coverage for overflowing millisecond duration values.

Non-goals:
- Do not change duration string parsing through `time.ParseDuration`.
- Do not change generic integer or float field handling.

Files:
- internal/config/manager.go
- internal/config/config_test.go

Tests:
- go test -timeout 20s ./internal/config
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; internal unmarshal hardening only.

Done Definition:
- Numeric millisecond duration fallback cannot wrap on multiplication.
- Overflow behavior is covered by tests and returns an error instead of panic or wrap.

Outcome:

