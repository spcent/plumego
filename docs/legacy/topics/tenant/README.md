# Tenant Docs (Historical Planning)

This folder primarily contains historical planning/checklist artifacts.
It is **not** the canonical source for v1 API usage.

For current guidance, use:

1. `docs/README.md` (canonical navigation and priority)
2. `docs/other/V1_GA_PRODUCTION_SCOPE.md` (GA vs experimental scope)
3. `docs/modules/tenant/README.md` and `docs/modules/tenant/*.md` (current tenant usage docs)
4. `README.md` / `README_CN.md` tenant snippets (canonical quick wiring)

Compatibility note:

- `tenant/*` is experimental in v1.0 (not GA-stable in v1.x guarantees).
- Some files in this directory contain pre-freeze option-based snippets
  (e.g. `WithTenantMiddleware`) that are kept for historical context only.
