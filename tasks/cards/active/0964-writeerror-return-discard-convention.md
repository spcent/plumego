# Card 0964: WriteError Return-Value Discard Convention

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: active
Primary Module: middleware
Owned Files: middleware/auth/contract.go, router/dispatch.go
Depends On: 0963

## Goal

Apply the canonical `_ = contract.WriteError(...)` pattern uniformly in middleware and
router code that currently calls `contract.WriteError` without discarding the return value.

## Problem

Style Guide §17.3 establishes one canonical error-write path:

```go
_ = contract.WriteError(w, r, err)
```

The `_ =` discard makes intent explicit: once any header is written, a second write has no
effect on the client, so the error return is genuinely unactionable at that call site.

Several stable-root files call `contract.WriteError` without the `_ =`, inconsistent with
the canonical pattern already used in extensions (e.g., `x/webhook/in.go`):

| File | Lines | Count |
|---|---|---|
| `middleware/auth/contract.go` | 129, 143, 151 | 3 |
| `router/dispatch.go` | 48 | 1 |

`middleware/debug/debug_errors.go` line 129 is covered by card 0963 (that card also
touches `debug_errors.go` for the config-merge fix).

## Scope

- Add `_ =` in front of the four bare `contract.WriteError(...)` calls listed above.
- No behavioral change.

## Non-Goals

- Do not change the error payload, category, or code in any of the affected calls.
- Do not touch `middleware/debug/debug_errors.go` — handled by card 0963.
- Do not sweep extension packages (`x/*`) in this card; they can be addressed as
  follow-up if the linter catches remaining instances.

## Files

- `middleware/auth/contract.go`
- `router/dispatch.go`

## Tests

```bash
go test -timeout 20s ./middleware/auth/... ./router/...
go vet ./middleware/auth/... ./router/...
```

## Docs Sync

None.

## Done Definition

- `middleware/auth/contract.go` lines 129, 143, 151 each begin `_ = contract.WriteError(`.
- `router/dispatch.go` line 48 begins `_ = contract.WriteError(`.
- Tests and vet pass.

## Outcome

