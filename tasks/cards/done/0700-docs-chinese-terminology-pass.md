# 0700 - Tighten Chinese Documentation Terminology

State: done
Priority: P2
Primary module: website docs

## Goal

Improve Chinese documentation polish by reducing unnecessary English labels and
making recurring terms consistent across high-traffic `/zh/docs` pages while
preserving package names, function names, and intentional technical identifiers.

## Scope

- Create or update a lightweight terminology convention inside the touched docs
  pages, not a large new glossary unless needed.
- Keep package names such as `core`, `router`, `contract`, `x/*`, and function
  names unchanged.
- Localize section labels and explanatory copy where English remains unnecessary.
- Keep Chinese pages aligned with English meaning.

## Non-goals

- Do not translate Go identifiers, import paths, HTTP methods, or public symbols.
- Do not rewrite every Chinese page in one pass.
- Do not change sidebar order or route slugs.

## Page Revision Checklist

| Target page | Modification goal | Acceptance points |
| --- | --- | --- |
| `website/src/content/docs/zh/docs/index.mdx` | Tighten homepage terminology for canonical path, stable surface, and repository facts. | English is kept only for intentional terms or package identifiers; section labels read naturally in Chinese. |
| `website/src/content/docs/zh/docs/getting-started.mdx` | Normalize "canonical", "first run", handler, and bootstrap wording. | The page stays clear for Chinese readers while preserving exact commands and API names. |
| `website/src/content/docs/zh/docs/reference-app.mdx` | Normalize reference app, bootstrap, route ownership, and response helper language. | The page no longer mixes avoidable English phrasing in section labels or table descriptions. |
| `website/src/content/docs/zh/docs/concepts/request-flow.mdx` | Normalize request-flow labels used by the visual and tables. | Package/function names remain unchanged; explanatory labels are Chinese. |
| `website/src/content/docs/zh/docs/concepts/repo-control-plane.mdx` | Normalize control-plane terminology. | `docs`, `specs`, `tasks`, and `module.yaml` remain identifiers; surrounding labels are Chinese. |
| `website/src/content/docs/zh/docs/modules/overview.mdx` | Normalize stable root, extension family, primary entrypoint, and subordinate package wording. | Recurring terms match the rest of the Chinese docs. |
| `website/src/content/docs/zh/docs/stable-roots.mdx` | Normalize stable-root explanation language. | The page distinguishes stable root, default path, and extension family without awkward mixing. |
| `website/src/content/docs/zh/docs/x-family.mdx` | Normalize extension-family wording. | Chinese copy keeps `x/*` and package names intact while localizing explanatory text. |
| `website/src/content/docs/zh/docs/release-posture.mdx` | Normalize release-posture and compatibility language. | Stable/experimental distinctions remain technically precise. |
| `website/src/content/docs/zh/docs/guides/index.mdx` | Normalize guide category labels and short descriptions. | Guide titles and descriptions read consistently with linked page terminology. |

## Files

- `website/src/content/docs/zh/docs/index.mdx`
- `website/src/content/docs/zh/docs/getting-started.mdx`
- `website/src/content/docs/zh/docs/reference-app.mdx`
- `website/src/content/docs/zh/docs/concepts/request-flow.mdx`
- `website/src/content/docs/zh/docs/concepts/repo-control-plane.mdx`
- `website/src/content/docs/zh/docs/modules/overview.mdx`
- `website/src/content/docs/zh/docs/stable-roots.mdx`
- `website/src/content/docs/zh/docs/x-family.mdx`
- `website/src/content/docs/zh/docs/release-posture.mdx`
- `website/src/content/docs/zh/docs/guides/index.mdx`

## Tests

- `cd website && pnpm check:docs-api` - passed
- `cd website && pnpm check` - passed
- `cd website && pnpm build` - passed

## Planned Tests

- `cd website && pnpm check`
- `cd website && pnpm build`

## Docs Sync

This is localization polish. Keep meaning aligned with English pages.

## Done Definition

- High-traffic Chinese docs pages avoid unnecessary English UI labels.
- Technical identifiers remain exact.
- Recurring translated terms are consistent across the touched pages.
