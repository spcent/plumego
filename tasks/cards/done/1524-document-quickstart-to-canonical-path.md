# Card 1524

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Context Package: control-plane
Priority: P2
State: done
Primary Module: docs/start
Owned Files:
- `docs/start/getting-started.md`
- `docs/start/getting-started_CN.md`
- `reference/standard-service/README.md`

## Goal

Bridge the gap between the three bootstrap forms shown in docs so new users
understand the progression and why `reference/standard-service` uses a different
shape than the quickstart examples.

## Scope

Docs-only additions to three files:

1. **`docs/start/getting-started.md`**: add a short "Progression" section (≤ 10
   lines) after the existing "Production-Style Canonical Example" block that names
   the three forms and their use-cases:
   - `plumego.New()` — zero-config hello-world
   - `core.New` + inline `Prepare/Server/Shutdown` — explicit lifecycle, logger
     injection (shown in this file)
   - `app.New` + `RegisterRoutes` + `App.Start` wrapper — canonical app layout,
     `reference/standard-service`

   One sentence each; no code snippets needed. End with: "When the inline example
   stops being enough, copy structure from `reference/standard-service` rather than
   extending `main.go`."

2. **`reference/standard-service/README.md`**: add one sentence to the first
   paragraph explaining that this service uses the third form (`app.New` wrapper)
   and links back to `docs/start/getting-started.md` for the progression context.

3. **`docs/start/getting-started_CN.md`**: apply the equivalent Chinese-language
   additions that mirror the English changes in item 1. Keep phrasing parallel;
   do not translate idiomatically beyond what is already present in the CN file.

Do not add a `routeReg` explanation to getting-started; that belongs only in
`ARCHITECTURE.md` and the style guide.

## Non-goals

- Do not change any Go source file.
- Do not add new code examples or new HTTP routes.
- Do not modify `ARCHITECTURE.md`, `AGENTS.md`, or `specs/`.
- Do not run `make website-sync`; none of the three files is a docs-sync source
  listed in AGENTS.md §7.

## Files

- `docs/start/getting-started.md`
- `docs/start/getting-started_CN.md`
- `reference/standard-service/README.md`

## Acceptance Tests

— (docs-only; verified by prose review)

## Tests

None required.

## Docs Sync

`docs/start/getting-started.md` and `docs/start/getting-started_CN.md` are not
listed as `make website-sync` sources in AGENTS.md §7; no regeneration needed.

## Validation

```
git diff --check
```

## Done Definition

- [x] `getting-started.md` contains a "Progression" section naming all three
  bootstrap forms in order.
- [x] `getting-started_CN.md` mirrors the English additions.
- [x] `reference/standard-service/README.md` first paragraph links to the progression
  context in `getting-started.md`.
- [x] `git diff --check` exits 0.
- [x] No Go source files are modified.

## Outcome

Commit `5aeddda`. Added a "Bootstrap Progression" section to
`docs/start/getting-started.md` (after the Production-Style Canonical Example)
naming the three forms — `plumego.New()`, `core.New` + inline lifecycle,
`app.New` + `RegisterRoutes` — with one sentence each and the closing
guidance to copy structure from `reference/standard-service` rather than
extending `main.go`. Mirrored the section in `getting-started_CN.md` and
updated the first paragraph of `reference/standard-service/README.md` to
identify itself as the third form with a back-link.

Validation: `git diff --check` clean; no Go files modified.
