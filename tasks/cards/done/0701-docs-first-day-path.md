# 0701 - Add First-Day Docs Path

State: done
Priority: P2
Primary module: website docs

## Goal

Add a concise "first day" or "first 30 minutes" path that tells new readers the
minimum commands to run, the exact success signals to expect, and the next three
pages to read before they open the full module catalog.

## Scope

- Keep the path short and task-oriented.
- Use current reference app commands and response envelopes.
- Add the path in English and Chinese.
- Prefer reusing the visual-anchor work from card 0696 if it already exists.

## Non-goals

- Do not create a marketing landing page.
- Do not duplicate the full Getting Started page.
- Do not add claims beyond the current reference app behavior.

## Page Revision Checklist

| Target page | Modification goal | Acceptance points |
| --- | --- | --- |
| `website/src/content/docs/docs/index.mdx` | Add a compact first-day path card near the top. | The card lists `go run ./reference/standard-service`, `/healthz`, and `/api/v1/greet?name=Alice`; it links to Getting Started, Reference App, and Request Flow. |
| `website/src/content/docs/zh/docs/index.mdx` | Add the localized first-day path card. | Chinese card uses the same commands and localized next-page links. |
| `website/src/content/docs/docs/getting-started.mdx` | Add a short "what you should have after 30 minutes" summary. | The summary names the running service, two verified requests, and the files to inspect next. |
| `website/src/content/docs/zh/docs/getting-started.mdx` | Add the localized summary. | Chinese summary stays concise and mirrors the English intent. |
| `website/src/content/docs/docs/faq.mdx` | Update "Where should I start?" to reference the first-day path. | The FAQ points readers to `/docs` first and avoids duplicating the full path. |
| `website/src/content/docs/zh/docs/faq.mdx` | Update the localized FAQ entry. | Chinese FAQ links to `/zh/docs` and references the localized first-day path. |
| `website/src/content/docs/docs/guides/index.mdx` | Add one sentence that guides are the next step after the first-day path. | The guide index does not become another overview page; it simply sets the prerequisite. |
| `website/src/content/docs/zh/docs/guides/index.mdx` | Add the localized prerequisite sentence. | Chinese guide index keeps the same entry condition. |

## Files

- `website/src/content/docs/docs/index.mdx`
- `website/src/content/docs/zh/docs/index.mdx`
- `website/src/content/docs/docs/getting-started.mdx`
- `website/src/content/docs/zh/docs/getting-started.mdx`
- `website/src/content/docs/docs/faq.mdx`
- `website/src/content/docs/zh/docs/faq.mdx`
- `website/src/content/docs/docs/guides/index.mdx`
- `website/src/content/docs/zh/docs/guides/index.mdx`

## Tests

- `cd website && pnpm check:docs-api` - passed
- `cd website && pnpm check` - passed
- `cd website && pnpm build` - passed

## Planned Tests

- `cd website && pnpm check:docs-api`
- `cd website && pnpm check`

## Docs Sync

This is docs-only work. Keep the command examples aligned with the current
reference app.

## Done Definition

- A new reader can identify the minimum command path from `/docs` without
  scanning the full docs tree.
- The path points to existing pages and current response shapes.
- English and Chinese pages stay aligned.
