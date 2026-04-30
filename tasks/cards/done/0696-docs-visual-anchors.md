# 0696 - Add Visual Anchors To Core Docs Pages

State: done
Priority: P2
Primary module: website docs

## Goal

Add a small set of high-signal visual anchors to the most important `/docs`
entry pages so readers can understand the documentation path, request flow,
service assembly, module ownership, and release posture without relying only on
paragraphs and tables.

## Scope

- Keep the current docs information architecture.
- Prefer reusable Astro components or Mermaid-style diagrams over hand-maintained
  screenshots.
- Keep diagrams aligned with current repository facts and canonical API names.
- Apply the same visual treatment to English and Chinese pages where the target
  page has a localized counterpart.

## Non-goals

- Do not redesign the full website theme.
- Do not add decorative images that do not explain a real Plumego concept.
- Do not change public Go APIs or documentation content beyond presentation and
  short support text needed by the visuals.

## Page Revision Checklist

| Target page | Modification goal | Acceptance points |
| --- | --- | --- |
| `website/src/content/docs/docs/index.mdx` | Add a compact "first 30 minutes" path visual above or near the existing first successful path. | The page shows the smallest route from evaluation to first verified handler; the visual includes run, verify, inspect, and next-read states; links still point to current `/docs` routes. |
| `website/src/content/docs/zh/docs/index.mdx` | Add the localized version of the same path visual. | Chinese copy is idiomatic; no untranslated labels except intentional package names; links point to `/zh/docs/...`. |
| `website/src/content/docs/docs/concepts/request-flow.mdx` | Upgrade the request-flow visual into a fuller route-to-response diagram. | The diagram shows route registration, router match, middleware chain, handler, and `contract.WriteResponse` / `contract.WriteError`; the page still preserves the current table as the audit checklist. |
| `website/src/content/docs/zh/docs/concepts/request-flow.mdx` | Localize the request-flow visual. | The visual uses Chinese labels and keeps Go package/function names unchanged. |
| `website/src/content/docs/docs/reference-app.mdx` | Add a service assembly map that shows `main.go`, `internal/app`, `internal/config`, and `internal/handler` as explicit ownership blocks. | The visual makes clear what is safe to copy and what should not be inferred from the reference app. |
| `website/src/content/docs/zh/docs/reference-app.mdx` | Add the localized service assembly map. | The same ownership model appears in Chinese, with localized explanatory labels. |
| `website/src/content/docs/docs/modules/overview.mdx` | Add a visual stable-root vs `x/*` ownership map. | The map makes it obvious that stable roots are the default path and `x/*` is optional capability work. |
| `website/src/content/docs/zh/docs/modules/overview.mdx` | Add the localized ownership map. | Chinese labels match the current terminology used by the page. |
| `website/src/content/docs/docs/release-posture.mdx` | Add a release-posture decision visual for stable roots, reference app, CLI, primary `x/*`, and subordinate `x/*`. | The visual reinforces adoption posture without overstating compatibility guarantees. |
| `website/src/content/docs/zh/docs/release-posture.mdx` | Add the localized release-posture visual. | Chinese copy preserves the experimental/stable distinction and links to localized posture pages. |

## Files

- `website/src/content/docs/docs/index.mdx`
- `website/src/content/docs/zh/docs/index.mdx`
- `website/src/content/docs/docs/concepts/request-flow.mdx`
- `website/src/content/docs/zh/docs/concepts/request-flow.mdx`
- `website/src/content/docs/docs/reference-app.mdx`
- `website/src/content/docs/zh/docs/reference-app.mdx`
- `website/src/content/docs/docs/modules/overview.mdx`
- `website/src/content/docs/zh/docs/modules/overview.mdx`
- `website/src/content/docs/docs/release-posture.mdx`
- `website/src/content/docs/zh/docs/release-posture.mdx`
- Optional new or updated shared component under `website/src/components/docs/`

## Tests

- `cd website && pnpm check:docs-api` - passed
- `cd website && pnpm check` - passed
- `cd website && pnpm build` - passed

## Planned Tests

- `cd website && pnpm check`
- `cd website && pnpm build`

## Docs Sync

This is docs-only work. Keep English and Chinese pages in parity.

## Done Definition

- Each target page has a visual anchor that explains a real decision or flow.
- No target page relies on decoration-only imagery.
- The visual labels and surrounding copy match the current repository boundaries.
- English and Chinese pages remain route-aligned and concept-aligned.
