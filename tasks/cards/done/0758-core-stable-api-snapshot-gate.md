# 0758 - Core Stable API Snapshot Gate

State: done
Priority: P0
Primary module: quality gates

## Goal

Make the checked-in core stable API snapshot enforceable by the main quality gate.

## Scope

- Add a stable API snapshot check entrypoint for `core`.
- Regenerate the current core API surface into a temporary file.
- Compare the generated file with `docs/stable-api/snapshots/core-head.snapshot`.
- Wire the check into `make gates`.

## Non-goals

- Do not expand the stable API surface.
- Do not regenerate snapshots in place.
- Do not add checks for every stable root in this card.

## Files

- `scripts/check-stable-api-snapshots.sh`
- `Makefile`
- `tasks/cards/active/0758-core-stable-api-snapshot-gate.md`

## Tests

- `bash scripts/check-stable-api-snapshots.sh`
- `make gates`

## Docs Sync

No user-facing docs update required; this is a quality gate change.

## Done Definition

- `make gates` compares the current core public API with the checked-in snapshot.
- The snapshot check passes without mutating tracked files.

## Validation

- `bash scripts/check-stable-api-snapshots.sh`
- `GOCACHE=/private/tmp/plumego-gocache make gates`
