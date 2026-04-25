# Card 0982: Replace `err ==` Sentinel Comparisons with `errors.Is()`

Priority: P2
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/fileapi, x/webhook, x/rest, x/data, x/mq, x/resilience

## Goal

Multiple files use direct equality `err == SentinelErr` to test for known error
values.  This pattern fails silently when an error is **wrapped** (e.g.
`fmt.Errorf("context: %w", sentinel)`), because the outer value does not equal
the inner sentinel.  Go best practice since 1.13 is to use `errors.Is()` which
unwraps the error chain.

## Affected Sites

### `x/fileapi/handler.go`

| Line | Current | Replace with |
|---|---|---|
| 313 | `if err == storefile.ErrNotFound {` | `if errors.Is(err, storefile.ErrNotFound) {` |

### `x/webhook/out.go`

| Line | Current | Replace with |
|---|---|---|
| 192 | `if err == ErrTargetNotFound {` | `if errors.Is(err, ErrTargetNotFound) {` |

### `x/rest/resource_db.go`

| Line | Current | Replace with |
|---|---|---|
| 480 | `if err == sql.ErrNoRows {` (Show) | `if errors.Is(err, sql.ErrNoRows) {` |
| 617 | `if err == sql.ErrNoRows {` (Update) | `if errors.Is(err, sql.ErrNoRows) {` |
| 676 | `if err == sql.ErrNoRows {` (Delete) | `if errors.Is(err, sql.ErrNoRows) {` |

### `x/data/file/metadata.go`

| Line | Current | Replace with |
|---|---|---|
| 73 | `if err == sql.ErrNoRows {` | `if errors.Is(err, sql.ErrNoRows) {` |
| 110 | `if err == sql.ErrNoRows {` | `if errors.Is(err, sql.ErrNoRows) {` |
| 148 | `if err == sql.ErrNoRows {` | `if errors.Is(err, sql.ErrNoRows) {` |

### `x/mq/persistence.go`

| Line | Current | Replace with |
|---|---|---|
| 129 | `if err == kvengine.ErrKeyNotFound \|\| err == kvengine.ErrKeyExpired {` | `if errors.Is(err, kvengine.ErrKeyNotFound) \|\| errors.Is(err, kvengine.ErrKeyExpired) {` |
| 215 | same pattern | same fix |
| 271 | same pattern | same fix |

### `x/resilience/circuitbreaker/middleware.go`

| Line | Current | Replace with |
|---|---|---|
| 40 | `if err == ErrCircuitOpen {` | `if errors.Is(err, ErrCircuitOpen) {` |
| 45 | `if err == ErrTooManyRequests {` | `if errors.Is(err, ErrTooManyRequests) {` |

## Notes

- Add `"errors"` import where not already present.
- Do **not** change `err == nil` checks — the nil sentinel does not need
  `errors.Is` and the Go `errors.Is(err, nil)` form is unusual.
- Do **not** change test-only comparisons (they are not in the target scope).

## Tests

```bash
go test -timeout 20s ./x/fileapi/... ./x/webhook/... ./x/rest/... ./x/data/... ./x/mq/... ./x/resilience/...
go vet ./x/fileapi/... ./x/webhook/... ./x/rest/... ./x/data/... ./x/mq/... ./x/resilience/...
```

## Done Definition

- No `err == <sentinel>` pattern (where sentinel is a non-nil exported error
  variable) appears in any of the listed files.
- All test suites pass; `go vet` is clean.

## Outcome

Completed. All `err == sentinel` comparisons in listed production files replaced with `errors.Is()`:
- `x/fileapi/handler.go`: ErrNotFound
- `x/webhook/out.go`: ErrTargetNotFound
- `x/rest/resource_db.go`: sql.ErrNoRows (3 sites)
- `x/data/file/metadata.go`: sql.ErrNoRows (3 sites)
- `x/mq/persistence.go`: kvengine.ErrKeyNotFound/ErrKeyExpired (3 sites)
- `x/resilience/circuitbreaker/middleware.go`: ErrCircuitOpen, ErrTooManyRequests
All tests pass; `go vet` clean.
