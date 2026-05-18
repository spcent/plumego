# Card 0192

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: website
Owned Files: website/src/lib/**, website/src/components/shared/**, website/src/components/starlight/**, website/src/components/docs/**, website/src/layouts/MarketingLayout.astro
Depends On: none

Goal:

Consolidate shared website locale metadata and routing helpers so shared layout,
SEO, and docs components do not encode two-language assumptions with scattered
`locale === 'zh'` branches.

Scope:

- Add a central locale metadata module for path prefixes, HTML language tags,
  Open Graph locale values, default OG images, language switch labels, and
  shared accessibility labels.
- Update shared i18n helpers to derive locale detection and localized paths from
  the locale metadata.
- Update SEO helpers and head/layout renderers to produce alternate locale data
  from configured locales instead of fixed English/Chinese fields.
- Replace shared component two-way locale branches with locale-indexed copy
  tables where the copy is still component-local.
- Keep existing English default routes and Chinese `/zh` routes unchanged.

Non-goals:

- Do not add a third locale.
- Do not rewrite marketing pages or MDX content into a new translation system.
- Do not change Starlight sidebar content strategy.
- Do not introduce new dependencies.
- Do not change Plumego Go package APIs.

Files:

- `website/src/lib/locales.ts`
- `website/src/lib/i18n.ts`
- `website/src/lib/seo.ts`
- `website/src/layouts/MarketingLayout.astro`
- `website/src/components/starlight/Head.astro`
- `website/src/components/starlight/Banner.astro`
- `website/src/components/shared/SiteHeader.astro`
- `website/src/components/shared/LanguageSelect.astro`
- `website/src/components/docs/*.astro` with locale copy branches

Tests:

- Add focused tests for locale detection, localized paths, locale switching, and
  SEO alternate generation if the existing website test stack can run them
  without new dependencies.

Docs Sync:

- Update website documentation only if the i18n rules or URL strategy changes.
  This card preserves the current URL strategy, so no docs sync is expected.

Done Definition:

- Shared locale metadata lives in one module and is used by i18n, SEO, headers,
  language switcher, and shared layout/head rendering.
- Shared `locale === 'zh'` branches are removed where a locale-indexed table or
  metadata lookup is the clearer extension point.
- Existing English and Chinese URLs, alternates, HTML language tags, and OG
  locale values remain unchanged.
- `pnpm check` passes in `website`.

Preflight:

- Owning module: website.
- Target module.yaml read: not applicable; `website` has no module manifest.
- In-scope paths: listed under Owned Files and Files.
- Out-of-scope paths: Go stable roots, `x/*`, generated sync outputs unless a
  validation command regenerates them, page copy rewrites, new locales.
- Public API impact: none for Go packages; internal website helper return shape
  changes are allowed inside this card.
- Dependency impact: none.
- Behavior impact: SEO/head generation and locale switch path generation should
  be behavior-preserving for current en/zh routes.
- Security impact: none.
- Docs impact: none expected because URL strategy is unchanged.
- Validation plan: run focused website checks and inspect remaining locale
  branch usage for shared components.

Outcome:

- Added `website/src/lib/locales.ts` as the central source for locale metadata,
  path prefixes, language tags, Open Graph locale values, default OG images, and
  shared accessibility labels.
- Reworked shared i18n and SEO helpers to derive localized paths, canonical
  URLs, alternate links, and alternate OG locales from configured locales.
- Updated marketing layout, Starlight head/banner overrides, shared header,
  language switcher, and reusable docs components to use locale metadata or
  locale-indexed copy tables instead of scattered two-language branches.
- Preserved the current URL strategy: English remains unprefixed and Chinese
  remains under `/zh`.
- Validation passed with `pnpm check` and `pnpm build` in `website`.
