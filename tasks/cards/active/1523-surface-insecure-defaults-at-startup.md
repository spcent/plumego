# Card 1523

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Context Package: implementation
Priority: P2
State: active
Primary Module: reference/standard-service
Owned Files:
- `reference/standard-service/internal/app/app.go`
- `reference/standard-service/internal/app/app_test.go`

## Goal

Emit a single `Warn`-level structured log line at startup for each insecure default
that is active, so operators can observe the risk without reading the checklist.

Two conditions to surface:

1. **Write guard disabled**: mutating routes are registered AND `cfg.App.WriteKey == ""`
   → log `"write guard disabled: POST/PUT/PATCH/DELETE /api/v1/items are publicly writable; set APP_WRITE_KEY in production"`.

2. **CORS wildcard**: `cfg.App.CORSAllowedOrigins` is empty
   → log `"CORS allows all origins (*); set APP_CORS_ALLOWED_ORIGINS in production"`.

Both warnings must be emitted from `App.New`, after middleware construction and
before returning, using `app.Logger().Warn(...)` with a `plumelog.Fields` map.

## Non-goals

- Do not change route registration, middleware wiring, or any other behavior in
  `app.go`; only add `Logger().Warn(...)` calls.
- Do not add a warning for `MaxBodyBytes == 0` (limit disabled is a valid
  intentional config; WriteKey and CORS wildcard are the highest-impact risks).
- Do not gate startup or return an error on these conditions — they are warnings,
  not hard failures. The reference must still start with default config.
- Do not touch any file outside `reference/standard-service/`.

## Files

- `reference/standard-service/internal/app/app.go`
- `reference/standard-service/internal/app/app_test.go`

## Acceptance Tests

```
reference/standard-service/internal/app/app_test.go: TestAcceptanceInsecureDefaultsWarnWriteKey
reference/standard-service/internal/app/app_test.go: TestAcceptanceInsecureDefaultsWarnCORSWildcard
reference/standard-service/internal/app/app_test.go: TestAcceptanceInsecureDefaultsNoWarnWhenConfigured
```

`TestAcceptanceInsecureDefaultsNoWarnWhenConfigured` verifies that no warning is
emitted when `WriteKey` is non-empty and `CORSAllowedOrigins` is set — to prevent
false-positive noise in correctly configured deployments.

Write all three functions first and confirm they **fail** before touching `app.go`.

## Tests

- Use `plumelog.NewLogger(plumelog.LoggerConfig{Format: plumelog.LoggerFormatJSON})`
  with a captured `bytes.Buffer` writer to assert that the warning message appears
  (or does not appear) in the log output. Do not assert on exact field order; assert
  on the presence of the message string and the `"level":"warn"` field.

## Docs Sync

- `reference/standard-service/PRODUCTION_CHECKLIST.md`: add a note under the
  WriteKey and CORS items that the service now logs a warning at startup when these
  defaults are active.

## Validation

```
cd reference/standard-service && go test -race -timeout 30s ./internal/app/...
go run ./internal/checks/reference-layout
git diff --check
```

## Done Definition

- [ ] Acceptance Tests pass (warn present when insecure, absent when configured).
- [ ] All Validation commands exit 0.
- [ ] `gofmt -l .` (inside `reference/standard-service`) produces no output.
- [ ] PRODUCTION_CHECKLIST.md updated to note the startup warning.
- [ ] `reference/standard-service` still starts cleanly with `go run .` and default config.

## Outcome

<!-- Agent fills this after completion. -->
