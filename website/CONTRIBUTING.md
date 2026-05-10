# Contributing to the Plumego Website

This document explains how to add or edit documentation, run a local preview, and meet the standards enforced by the content checkers.

## Quick start

Requirements: Node.js 20+, pnpm 10+.

```bash
cd website
pnpm install
pnpm dev        # starts dev server on http://localhost:4321
```

`pnpm dev` runs `pnpm sync` first to pull generated facts from the repository (roadmap, modules, releases, translation lag). You need the full repository checked out — the sync scripts read files from the parent directory.

---

## Adding a new doc page

### 1. Choose the right location

| Content type | Directory | Examples |
| --- | --- | --- |
| Concept explanation | `src/content/docs/docs/concepts/` | `error-model.mdx` |
| How-to guide | `src/content/docs/docs/guides/` | `add-jwt-auth.mdx` |
| Module primer | `src/content/docs/docs/modules/` | `core.mdx` |
| API quick reference | `src/content/docs/docs/reference/` | `api-core.mdx` |
| Top-level info | `src/content/docs/docs/` | `faq.mdx` |

Create the Chinese equivalent in the matching path under `src/content/docs/zh/`.

### 2. Required frontmatter

Every `.mdx` file must include:

```yaml
---
title: Short title (used in sidebar and <title>)
description: One sentence — shown in search results and social previews.
translationKey: unique-slug-matching-the-en-counterpart
sidebar:
  order: <integer>   # controls position within the sidebar group
---
```

`translationKey` links the en and zh versions for translation-lag detection. It must be identical in both files and unique across the entire docs tree.

### 3. Content rules (short version)

1. **Code before prose** on every feature and module page (CONTENT_RULES.md Rule 1).
2. **Stability badge first** on every `x/*` module page — use the `> **Experimental** ...` blockquote pattern.
3. **No circular navigation** — end each page with a "Read next" list of 2–5 links; do not duplicate the AdoptionNav four-card block.
4. **No invented performance claims** — any latency or throughput number must link to a reproducible benchmark (CONTENT_RULES.md Rule 8).
5. **Guide pages must open with a runnable terminal command** (CONTENT_RULES.md Rule 10).

Full rules: [`CONTENT_RULES.md`](./CONTENT_RULES.md).

### 4. Register the page in the sidebar

Open `astro.config.mjs` and add an entry in the correct sidebar group:

```js
{ label: 'My New Page', slug: 'docs/guides/my-new-guide', translations: { 'zh-CN': '我的新指南' } },
```

The `slug` must match the file path relative to `src/content/docs/` and without the `.mdx` extension.

---

## Editing an existing page

1. Edit the English file under `src/content/docs/docs/`.
2. Update the matching Chinese file under `src/content/docs/zh/docs/` in the same commit when the change is non-trivial, or open a follow-up issue if the translation can wait.
3. If the Chinese version is intentionally left behind, the translation-lag checker will surface it automatically after 14 days.

---

## Running the content checkers

```bash
pnpm check:docs-api   # verifies API symbols referenced in docs exist in the Go source
pnpm check:css        # verifies CSS selectors are not broken
pnpm check            # runs sync + both checks + Astro type-check
```

The `prebuild` and `precheck` scripts run `pnpm sync` automatically. Run `pnpm sync` manually to regenerate facts after changing specs, README, or roadmap.

---

## Translation lag detection

The sync pipeline runs `scripts/check-translation-lag.mjs` at build time. It compares the git commit timestamp of each en file with its zh counterpart. Pages where the zh version is more than **14 days** behind generate an entry in `src/generated/translation-lag.ts`.

zh pages that import `TranslationNotice` can render a visible banner:

```mdx
import TranslationNotice from '../../../components/docs/TranslationNotice.astro';

<TranslationNotice translationKey="getting-started" />
```

To add the notice to an existing zh page, import the component and add it immediately after the `# Title` heading.

---

## Running a production build

```bash
pnpm build    # outputs to dist/
pnpm preview  # serves dist/ locally
```

The build target is Cloudflare Pages (static output). Do not add server-side rendering; the site is fully static.

---

## Generated files

Files in `src/generated/` are created by the sync scripts. Do not edit them by hand — they are overwritten on every `pnpm sync` run.

| Generated file | Source | Script |
| --- | --- | --- |
| `modules.ts` | `specs/repo.yaml`, `specs/task-routing.yaml` | `sync-modules.mjs` |
| `releases.ts` | `README.md` | `sync-release-meta.mjs` |
| `roadmap.ts` | `docs/ROADMAP.md` | `sync-roadmap.mjs` |
| `translation-lag.ts` | git log of `src/content/docs/` | `check-translation-lag.mjs` |

---

## Editorial governance

Before adding new content, read [`EDITORIAL.md`](./EDITORIAL.md). The eight standing rules prevent content sprawl and keep each page answering one primary question.
