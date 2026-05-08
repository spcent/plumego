# Card 0732

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/recovery/recover.go
- middleware/recovery/recover_test.go
- docs/modules/middleware/README.md
Depends On:
- 0731-middleware-responsewriter-conformance-suite

Goal:
Prevent recovery middleware from logging raw panic payloads while preserving
request correlation and stable 500 behavior.

Scope:
- Replace raw panic-value logging with a sanitized panic classification.
- Keep client responses generic and unchanged.
- Update tests that currently assert raw panic details in logs.

Non-goals:
- Do not add stack trace collection.
- Do not change recovery response shape.
- Do not add dependencies.

Files:
- middleware/recovery/recover.go
- middleware/recovery/recover_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/recovery
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Raw panic strings cannot appear in recovery log fields.
- Recovery still logs method, path, status, duration, and request ID.
- Targeted and middleware-wide tests pass.

Outcome:
- Recovery now logs `panic_type` instead of the raw recovered panic value.
- Added regression coverage proving raw panic strings are not present in log
  fields.
- Documented that recovery logs sanitized panic metadata only.

Validation:
- `go test -timeout 20s ./middleware/recovery`
- `go test -timeout 20s ./middleware/...`
