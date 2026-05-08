# Website Content Rules

Rules for writing and reviewing website content. Apply to all pages under `src/pages/` and `src/content/docs/`.

---

## Rule 1 — Code before prose

Every feature or module page must show a runnable code example as its **first content block**. Do not open with a conceptual description. If the page cannot show a meaningful code example, it may open with a CLI command or terminal output instead.

**Applies to:** module primers, guides, getting-started.

---

## Rule 2 — One page, one next step

Each documentation page must end with exactly **one** clear "next step" link. Do not offer three parallel options at the end of a page. If branching is necessary, use a table labeled "leave this page when" with a question in each row and one corresponding link.

---

## Rule 3 — Stability badge required on every x/* module page

Every page under `docs/modules/x-*.mdx` must display its stability tier as the **first visible element** after the title. Use one of:

- `**Stable**` — long-term API target, part of v1 hardening scope
- `**Experimental**` — API not frozen, compatibility not guaranteed

Do not bury maturity information in body text.

---

## Rule 4 — No internal terminology without definition

The following terms must be **defined on first use** on any user-facing page:

| Term | One-line definition to include |
| --- | --- |
| stable root | A module with a long-term API compatibility promise, part of the v1 hardening scope |
| x/* family | An optional capability module under the `x/` directory; experimental unless labeled otherwise |
| control plane | The `docs/`, `specs/`, and `tasks/` directories that hold human-readable rules and machine-readable constraints |
| canonical path | The default service layout defined by `reference/standard-service` |

Do not use these terms as if they are universally understood jargon.

---

## Rule 5 — Competitor comparisons must include "when not to choose Plumego"

Any page that compares Plumego with Gin, Echo, Chi, or other toolkits must include a column or section explicitly stating **when to choose the alternative** over Plumego. A comparison that only argues in favor of Plumego is not acceptable.

---

## Rule 6 — Chinese content sync transparency

If a `/zh/` page is known to lag its English counterpart, add the following notice at the top of the page body (after the frontmatter):

```md
> **注意：** 本页中文版本与英文版可能存在差异，建议关键决策前参考[英文版文档](/docs/...)。
```

The sync scripts (`website/scripts/sync-all.mjs`) should record a timestamp; any `/zh/` page more than 14 days behind the corresponding `/en/` page should display this notice automatically.

---

## Rule 7 — Homepage must be scannable in under 10 seconds

The homepage hero section must convey, **without scrolling**:

1. What Plumego is (one sentence)
2. A runnable code example (≤ 20 lines)
3. The primary call to action ("Get Started → /docs/getting-started")

Do not add marketing sections between the hero and the first code example.

---

## Rule 8 — Agent-first content is isolated to Advanced docs

Content about AI agent workflows, `specs/task-routing.yaml`, `specs/change-recipes/`, and machine-readable repository rules belongs **only** in:

- `docs/concepts/repo-control-plane`
- `docs/concepts/agent-first-workflow`

It must **not** appear on:

- Homepage (`src/pages/index.astro`)
- Why Plumego (`src/pages/why-plumego.astro`)
- Getting Started (`src/content/docs/docs/getting-started.mdx`)
- Any module primer

The intended audience for agent-first content is contributors and teams integrating AI tooling into their workflow — not developers evaluating Plumego for the first time.

---

## Rule 9 — JourneyBar consistency

Every marketing page that renders a `<JourneyBar>` must use the same three-step sequence:

```js
[
  { label: 'Why Plumego', href: '/why-plumego' },
  { label: 'Examples', href: '/examples' },
  { label: 'Releases', href: '/releases' },
]
```

Do not add, remove, or reorder steps between pages. The JourneyBar must visually convey a single adoption path, not a sitemap. Architecture is not a step in the adoption path — it is a reference page.

---

## Rule 10 — Examples pages must open with a runnable command

Any page whose title includes "Examples" or whose primary purpose is to show code must present a runnable terminal command or code block as its **first content block** after the hero/JourneyBar, before any descriptive prose or recipe lists.

The command must be concise enough to run in under 30 seconds: typically `go run ./reference/standard-service` plus one verification `curl`.

---

## Rule 11 — Pages without a translation must say so

If an English documentation page has no corresponding `/zh/` page, the English page must display the following notice in a blockquote at the start of the page body (after frontmatter):

```md
> **Note:** This page is not available in Chinese. Refer to the English version.
```

If a `/zh/` page exists but is known to lag its English counterpart by more than 14 days, apply Rule 6 instead.

---

## Rule 12 — One primary responsibility per page

Each documentation or marketing page must have exactly one primary responsibility stated in its frontmatter `description`. If a page answers more than two unrelated question types, split it or redirect secondary questions to the page that owns them.

Responsibility overlap test: if removing a section from the page would not break the page's primary purpose, the section probably belongs elsewhere.

**Examples of violations:**
- The getting-started page explaining module ownership decisions (belongs in Modules Overview)
- The modules overview page explaining how to run an example (belongs in Getting Started)
- The FAQ repeating information that is already the primary content of a guide page (keep the FAQ answer short and link to the guide)
