# Card 0742

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- middleware/accesslog/accesslog.go
- middleware/accesslog/accesslog_test.go
- docs/modules/middleware/README.md
Depends On:
- 0741-middleware-observability-safe-finalizers

Goal:
Use the shared middleware redaction policy for access-log fields.

Scope:
- Pass accesslog fields through `internalobs.RedactFields` before logging.
- Add regression coverage for sensitive field names in log output.
- Keep existing accesslog field names and values when not sensitive.

Non-goals:
- Do not add response/header/body logging.
- Do not change logger interfaces.
- Do not change accesslog observer/tracer wiring.

Files:
- middleware/accesslog/accesslog.go
- middleware/accesslog/accesslog_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/accesslog
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Accesslog uses the same redaction policy as other middleware logs.
- Sensitive field-name regression coverage exists.
- Targeted and middleware-wide tests pass.

Outcome:

