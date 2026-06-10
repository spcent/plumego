# Card 1520

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: control-plane
Priority: P1
State: active
Primary Module: reference/standard-service
Owned Files:
- `reference/standard-service/PRODUCTION_CHECKLIST.md`
- `reference/standard-service/internal/app/app.go`

## Goal

Fix three documentation inconsistencies in `reference/standard-service` that cause
agents and new users to reference non-existent identifiers or contradictory pointers.

## Scope

Three targeted prose corrections, docs-only:

1. `PRODUCTION_CHECKLIST.md:41` — replace `midsecurity.Config{}` with
   `securityheaders.Config{}` (the alias `midsecurity` does not exist in the codebase).

2. `PRODUCTION_CHECKLIST.md:44` — replace the ambiguous two-package pointer
   (`security/abuse` or `x/resilience/ratelimit`) with a single canonical choice plus
   a cross-reference: `middleware/abuseguard` is the stable-root option; point to
   `x/resilience/ratelimit` only for extension use-cases, matching the comment already
   in `internal/app/app.go:69-71`.

3. `internal/app/app.go:75` — replace `reference/with-observability` with
   `reference/with-ops` to match the authoritative metrics wiring example cited in
   `PRODUCTION_CHECKLIST.md:81`.

## Non-goals

- Do not change any Go source behavior, imports, or middleware wiring.
- Do not touch `README.md`, `ARCHITECTURE.md`, or `AGENT_TASKS.md`.
- Do not run `make website-sync`; these files are not docs-sync sources.

## Files

- `reference/standard-service/PRODUCTION_CHECKLIST.md`
- `reference/standard-service/internal/app/app.go`

## Acceptance Tests

— (docs-only; no behavior change; verified by `git diff --check` and prose review)

## Tests

None required. Confirm with `git diff --check` that no whitespace errors were introduced.

## Docs Sync

Not applicable. `PRODUCTION_CHECKLIST.md` and inline comments are not listed as
`make website-sync` sources.

## Validation

```
git diff --check
cd reference/standard-service && go test -race -timeout 30s ./...
```

## Done Definition

- [ ] `midsecurity.Config{}` does not appear anywhere under `reference/standard-service/`.
- [ ] Rate-limit guidance in `PRODUCTION_CHECKLIST.md` and `app.go` comments names the same package.
- [ ] Metrics example cross-reference in `app.go:75` matches `PRODUCTION_CHECKLIST.md:81`.
- [ ] All Validation commands exit 0.
- [ ] `gofmt -l .` (inside `reference/standard-service`) produces no output.

## Outcome

<!-- Agent fills this after completion. -->
