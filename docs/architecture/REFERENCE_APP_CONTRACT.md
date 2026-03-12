# Reference App Contract

`reference/standard-service` is the canonical application layout for Plumego.

## Purpose

- Define the one true project structure for docs and scaffolds.
- Provide a stable smoke target for repository validation.
- Keep the default developer path anchored to stable packages only.

## Rules

- The standard reference app uses only stable root packages by default.
- `x/*` packages are opt-in and must be wired explicitly.
- Templates must mirror the reference app layout.
- Examples must not introduce alternative architectural conventions.

## Current Layout

```text
reference/standard-service/
  main.go
  internal/app/
  internal/config/
  internal/docsite/
  internal/handler/
  ui/
```

## Canonical Expectations

- Explicit route registration.
- Standard-library handler signatures.
- Constructor-based dependency injection.
- `contract.WriteError` as the single error write path.
- No root package facade imports.
- Future templates must mirror this baseline until the reference app itself changes.
