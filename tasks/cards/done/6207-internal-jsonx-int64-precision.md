# 6207 - internal/jsonx int64 precision

State: done
Priority: P0
Primary module: `internal/jsonx`

## Goal

Preserve exact JSON integer values when `internal/jsonx` extracts int64 arrays
and maps. Current untyped decoding turns JSON numbers into `float64`, which can
round large identifiers before they are converted to `int64`.

## Scope

- Decode int64-oriented extractor inputs with `json.Decoder.UseNumber`.
- Keep existing best-effort behavior for supported numeric strings.
- Cover top-level and nested int64 collection helpers.

## Non-goals

- Do not change the public function names or return shapes.
- Do not rewrite the broader `jsonx` helper family.
- Do not change int or float extractors in this card.

## Files

- `internal/jsonx/json.go`
- `internal/jsonx/json_test.go`

## Tests

- `go test ./internal/jsonx`
- `go test ./internal/...`

## Docs Sync

No docs sync expected; this is internal behavior hardening without public API
shape changes.

## Done Definition

- Large int64 JSON numbers are returned exactly by array, map, array-map, and
  nested array-map extractors.
- Existing string-number behavior remains covered.
- Targeted and internal package tests pass.

## Outcome

- Added shared `UseNumber` decoding for int64 collection extractors.
- Preserved exact `MaxInt64` values in array, map, array-map, and nested
  array-map helpers.
- Validation:
  - `go test ./internal/jsonx`
  - `go test ./internal/...`
