# 0647 - internal/pool int64 precision

State: done
Priority: P0
Primary module: `internal/pool`

## Goal

Preserve exact JSON integer values when pooled extraction helpers return int64
slices and maps. Current untyped decoding can round large JSON numbers through
`float64`.

## Scope

- Decode int64-oriented pooled extractor inputs with `json.Decoder.UseNumber`.
- Keep existing best-effort behavior for numeric strings.
- Cover slice, map, and array-map int64 helpers.

## Non-goals

- Do not change exported helper names or return shapes.
- Do not rewrite pooled extraction helpers unrelated to int64.
- Do not introduce new pooling primitives.

## Files

- `internal/pool/pool.go`
- `internal/pool/pool_test.go`

## Tests

- `go test ./internal/pool`
- `go test ./internal/...`

## Docs Sync

No docs sync expected; this is internal helper hardening.

## Done Definition

- Large int64 JSON numbers are extracted exactly by pooled slice, map, and
  array-map helpers.
- Existing string-number behavior remains covered.
- Targeted and internal package tests pass.

## Outcome

- Added `UseNumber` decoding for pooled int64 JSON extractors.
- Preserved exact `MaxInt64` values in slice, map, and array-map helpers.
- Validation:
  - `go test ./internal/pool`
  - `go test ./internal/...`
