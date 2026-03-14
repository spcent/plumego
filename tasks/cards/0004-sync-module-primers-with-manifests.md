# Card 0004

Priority: P1

Goal:
- Ensure manifest-declared module primers exist and user-facing modules point to the right primer locations.

Scope:
- missing or misaligned `doc_paths`
- primer coverage for user-facing stable roots and extensions

Non-goals:
- Do not rewrite long-form tutorials.
- Do not add runtime behavior changes.

Files:
- `docs/modules/*`
- `core/module.yaml`
- `router/module.yaml`
- `x/webhook/module.yaml`
- `x/websocket/module.yaml`

Tests:
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Keep primer paths in manifests synchronized with actual docs.

Done Definition:
- Every manifest-declared primer path resolves to a real file.
- User-facing modules have a clear first-read primer.
- Primer locations no longer drift from manifest metadata.
