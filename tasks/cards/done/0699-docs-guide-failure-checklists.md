# 0699 - Add Failure Checklists To High-Traffic Guides

State: done
Priority: P2
Primary module: website docs

## Goal

Improve guide copyability by adding concise failure checklists to high-traffic
how-to pages. Each checklist should tell readers what to verify when a copied
example does not compile, does not route, or does not return the expected
contract envelope.

## Scope

- Start with guides that users are most likely to copy directly into services.
- Keep each checklist short and tied to current public APIs.
- Mirror English and Chinese pages.
- Prefer checks that can be performed by reading code or running one command.

## Non-goals

- Do not turn guides into troubleshooting manuals.
- Do not add new runtime examples that are not backed by current APIs.
- Do not change module ownership rules.

## Page Revision Checklist

| Target page | Modification goal | Acceptance points |
| --- | --- | --- |
| `website/src/content/docs/docs/guides/build-rest-resource.mdx` | Add a "If this does not work" checklist after the custom controller or production adoption section. | Checklist covers `x/rest` experimental posture, `rest.NewQueryBuilder().WithPageSize(...).Parse(r)`, route registration, repository method availability, and response envelope expectations. |
| `website/src/content/docs/zh/docs/guides/build-rest-resource.mdx` | Add the localized REST guide failure checklist. | Chinese checklist uses current API names and clearly distinguishes app-local controller issues from `x/rest` package issues. |
| `website/src/content/docs/docs/guides/handle-errors.mdx` | Add a short checklist for error responses that have the wrong shape or status. | Checklist verifies `NewErrorBuilder().Type(...)`, `WriteError`, `return` after write, and request-id middleware expectations. |
| `website/src/content/docs/zh/docs/guides/handle-errors.mdx` | Add the localized error guide checklist. | Chinese text preserves canonical function names and error type names. |
| `website/src/content/docs/docs/guides/add-jwt-auth.mdx` | Add an auth failure checklist. | Checklist covers token verification, fail-closed behavior, middleware order, and avoiding secret logging. |
| `website/src/content/docs/zh/docs/guides/add-jwt-auth.mdx` | Add the localized auth checklist. | Chinese checklist keeps security guidance explicit and conservative. |
| `website/src/content/docs/docs/guides/connect-database.mdx` | Add a database wiring failure checklist. | Checklist covers explicit dependency injection, context use, connection setup, and avoiding hidden globals. |
| `website/src/content/docs/zh/docs/guides/connect-database.mdx` | Add the localized database checklist. | Chinese wording keeps app-local wiring distinct from stable `store` primitives. |
| `website/src/content/docs/docs/guides/custom-middleware.mdx` | Add a middleware ordering checklist. | Checklist covers `func(http.Handler) http.Handler`, registration before `Prepare`, transport-only scope, and calling `next.ServeHTTP`. |
| `website/src/content/docs/zh/docs/guides/custom-middleware.mdx` | Add the localized middleware checklist. | Chinese text keeps the transport-only boundary clear. |
| `website/src/content/docs/docs/guides/testing-handlers.mdx` | Add a test failure checklist. | Checklist covers `httptest.NewRecorder`, response envelope assertions, status code assertions, and request context setup. |
| `website/src/content/docs/zh/docs/guides/testing-handlers.mdx` | Add the localized testing checklist. | Chinese checklist is concise and copy-focused. |

## Files

- `website/src/content/docs/docs/guides/build-rest-resource.mdx`
- `website/src/content/docs/zh/docs/guides/build-rest-resource.mdx`
- `website/src/content/docs/docs/guides/handle-errors.mdx`
- `website/src/content/docs/zh/docs/guides/handle-errors.mdx`
- `website/src/content/docs/docs/guides/add-jwt-auth.mdx`
- `website/src/content/docs/zh/docs/guides/add-jwt-auth.mdx`
- `website/src/content/docs/docs/guides/connect-database.mdx`
- `website/src/content/docs/zh/docs/guides/connect-database.mdx`
- `website/src/content/docs/docs/guides/custom-middleware.mdx`
- `website/src/content/docs/zh/docs/guides/custom-middleware.mdx`
- `website/src/content/docs/docs/guides/testing-handlers.mdx`
- `website/src/content/docs/zh/docs/guides/testing-handlers.mdx`

## Tests

- `cd website && pnpm check:docs-api` - passed
- `cd website && pnpm check` - passed
- `cd website && pnpm build` - passed

## Planned Tests

- `cd website && pnpm check:docs-api`
- `cd website && pnpm check`

## Docs Sync

This is docs-only work. Do not introduce symbols that are absent from the current
Go packages.

## Done Definition

- Each target guide has a concise failure checklist.
- Checklists mention only current public symbols and current response shapes.
- English and Chinese pages remain conceptually aligned.
