# Plumego Website

This directory contains the official Plumego website.

## Goals

- fully static output
- deployable to Cloudflare Pages
- Markdown/MDX-first publishing workflow
- bilingual support for English and Chinese
- shared light/dark theme across marketing pages and docs

## Stack

- Astro
- Starlight
- MDX
- Cloudflare Pages

## URL Strategy

- English is the default locale and has no prefix
- Chinese uses the `/zh` prefix

Examples:

- `/`
- `/docs/getting-started`
- `/roadmap`
- `/zh`
- `/zh/docs/getting-started`
- `/zh/roadmap`

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the full guide on how to add pages, required frontmatter, content rules, running checkers, and the translation lag workflow.

## Directory Rules

- `src/pages/**` owns marketing pages and top-level landing pages
- `src/content/docs/**` owns docs content
- `src/data/**` owns structured site metadata
- `public/**` owns static assets and deployment control files
- `scripts/**` owns build-time sync scripts only

Do not put runtime business logic here.
Do not import Plumego Go packages into the website build.
Do not couple the website to Go module internals.

## Content Rules

- Prefer MDX for publishable content
- Keep English and Chinese content in separate files
- Keep slugs aligned across locales
- Use frontmatter consistently
- Store facts and navigation metadata in `src/data/**`
- Sync only fact-like content from the repo:
  - roadmap summary
  - module inventory
  - release metadata

Do not mirror the full repository README into the website.

## Theme Rules

- default to system theme
- allow manual light/dark toggle
- persist user choice with the `starlight-theme` storage key
- use CSS tokens only
- support light/dark logo variants

## i18n Rules

- default locale: `en`
- secondary locale: `zh`
- locale switch maps equivalent pages when both exist
- when a translation does not exist, link back to the English page and label it clearly

## Deploy

Build output is emitted to `dist/`.

Deploy target:

- Cloudflare Pages

## Release Architecture

The website is split into two page families:

- marketing pages under `src/pages/**`
- documentation pages under `src/content/docs/**`

### Published navigation

Header navigation for the release-ready site is intentionally narrow:

- Docs
- Use Cases
- Examples
- Roadmap
- GitHub

Footer navigation is grouped by job instead of presented as one flat link row:

- Product: Docs, Use Cases, Examples
- Start: Getting Started, Reference App, FAQ
- Status: Roadmap, Releases, GitHub

### Page responsibilities

- `/` and `/zh`
  - product landing pages
  - explain positioning, value, adoption paths, and the canonical start
- `/docs` and `/zh/docs`
  - reading-first documentation entry pages
  - route readers through recommended path, topic groups, and synced repo facts
- `/docs/getting-started`
  - first-run path only
  - should answer “how do I successfully start?”
- `/docs/concepts/request-flow`, `/docs/concepts/repo-control-plane`, `/docs/faq`
  - concept-layer docs
  - should answer “how should I think about the request path and repository before changing code?”
- `/docs/reference-app`
  - canonical application layout
  - should answer “where do I copy the default service shape from?”
- `/docs/modules/overview`, `/docs/stable-roots`, `/docs/x-family`
  - ownership and boundary pages
  - should answer “where does a change belong?”
- `/docs/release-posture`
  - compatibility and maturity framing
  - should answer “how stable is this surface?”
- `/use-cases`
  - adoption-fit page
  - should answer “is Plumego a fit for my team and service shape?”
- `/examples`
  - example-path page
  - should answer “which example path should I trust first?”
- `/roadmap`
  - repository direction and explicit non-goals
- `/releases`
  - current version and support posture

### Release sitemap

The site currently ships every page under both `/` (English) and `/zh` (Chinese). The lists below are locale-agnostic — prepend `/zh` for the Chinese variant, except `/404` which has its own locale-specific files.

Marketing (`src/pages/**`):

- `/`
- `/why-plumego`
- `/use-cases`
- `/examples`
- `/extensions`
- `/architecture`
- `/agent-workflow`
- `/migrate`
- `/compare`
- `/status`
- `/stability`
- `/releases`
- `/roadmap`
- `/404` (English at `src/pages/404.astro`, Chinese at `src/pages/zh/404.astro`)

Docs entry (`src/content/docs/docs/*.mdx`):

- `/docs`
- `/docs/getting-started`
- `/docs/reference-app`
- `/docs/faq`
- `/docs/release-posture`
- `/docs/stable-roots`
- `/docs/x-family`
- `/docs/when-not-to-use`

Docs concepts (`src/content/docs/docs/concepts/*.mdx`):

- `/docs/concepts/agent-first-workflow`
- `/docs/concepts/configuration-model`
- `/docs/concepts/core-boundaries`
- `/docs/concepts/error-model`
- `/docs/concepts/extension-boundaries`
- `/docs/concepts/extension-maturity`
- `/docs/concepts/middleware-model`
- `/docs/concepts/repo-control-plane`
- `/docs/concepts/request-flow`

Docs guides (`src/content/docs/docs/guides/*.mdx`) — 22 files covering JWT auth, REST resources, database connection, middleware, Docker deploy, dev server, file uploads, graceful shutdown, error handling, health & readiness, AI integration, migration from chi/gin/echo, multi-tenancy, observability, structured logging, style guide, handler testing, WebSocket, and more.

Docs modules (`src/content/docs/docs/modules/*.mdx`) — 32 files:

- 10 stable roots: `contract`, `core`, `health`, `log`, `metrics`, `middleware`, `overview`, `router`, `security`, `store`
- 14 x/* families + 8 subordinate primers: `x-ai`, `x-cache`, `x-data`, `x-devtools`, `x-discovery`, `x-fileapi`, `x-frontend`, `x-gateway`, `x-ipc`, `x-messaging`, `x-messaging-subordinates`, `x-observability`, `x-openapi`, `x-ops`, `x-resilience`, `x-rest`, `x-rpc`, `x-scheduler`, `x-tenant`, `x-validate`, `x-webhook`, `x-websocket`

Docs reference (`src/content/docs/docs/reference/*.mdx`):

- `/docs/reference` (index)
- `/docs/reference/api-contract`
- `/docs/reference/api-core`
- `/docs/reference/api-router`
- `/docs/reference/deprecation`
- `/docs/reference/errors`
- `/docs/reference/stability`

Release-completion support:

- locale-aware canonical and Open Graph metadata on marketing and docs pages
- shared OG assets under `public/brand/**`
- bilingual EN/ZH parity checked by `scripts/check-translation-lag.mjs`

### Pages deployment shape

- framework preset: `Astro`
- project root: `website`
- build command: `pnpm build`
- output directory: `dist`

### Notes

- this website is a fully static build
- deployment is expected to be handled directly in Cloudflare Pages
- no Wrangler config is kept in the repository

## Cloudflare Pages Setup

Use the Cloudflare dashboard to connect this repository directly to Pages.

### 1. Create the Pages project

In Cloudflare Pages:

1. Go to `Workers & Pages` and create a new Pages project.
2. Connect the GitHub repository for Plumego.
3. Select the production branch.

Recommended production branch:

- `main`

### 2. Configure the build

Use these build settings:

- Framework preset: `Astro`
- Root directory: `website`
- Build command: `pnpm build`
- Build output directory: `dist`

If Cloudflare asks for the install command, use the default package-manager install flow or set:

- Install command: `pnpm install --frozen-lockfile`

### 3. Preview deployments

Cloudflare Pages will automatically create preview deployments for non-production branches and pull requests.

Important behavior:

- preview deployments are enabled by default
- the production branch updates your main `*.pages.dev` site
- other branches receive preview URLs
- pull request preview URLs are only guaranteed when the pull request originates from the same repository

Cloudflare also creates branch aliases for previews. For example, a branch like `feature/docs-home` will get a branch-style preview alias based on the branch name.

### 4. Environment variables

This website is currently a static build and does not require any custom environment variables to build successfully.

If you add build-time variables later, configure them in:

- Pages project → `Settings` → `Environment variables`

Cloudflare Pages also injects system variables such as:

- `CF_PAGES`
- `CF_PAGES_BRANCH`
- `CF_PAGES_URL`
- `CF_PAGES_COMMIT_SHA`

These can be useful later if the website needs branch-aware banners, preview-only behavior, or deployment metadata.

### 5. Custom domains

After the first successful production deployment:

1. Open the Pages project.
2. Go to `Custom domains`.
3. Add the desired domain or subdomain.

Common choices:

- production: `plumego.birdor.dev`

If you want a branch-specific custom domain later, Cloudflare Pages also supports attaching a custom domain to a branch alias such as `staging.example.com`.

### 6. Rollbacks

Production rollback is handled in the Cloudflare Pages dashboard:

- open the project
- go to `Deployments`
- choose a previous successful production deployment
- rollback from the deployment actions menu

Preview deployments are not rollback targets; rollback applies to production deployments.

### 7. Recommended Pages settings

After the project is live, review these settings in the dashboard:

- Production branch control
- Preview deployment access policy
- Custom domains
- Environment variables

If preview URLs should not be public, protect them with Cloudflare Access from the Pages project settings.

## Release Readiness Scope

Currently shipped (v1.1.0):

- marketing pages: home, why-plumego, use-cases, examples, extensions, architecture, agent-workflow, migrate, compare, status, stability, releases, roadmap, 404
- docs entry pages: docs home, getting-started, reference-app, faq, release-posture, stable-roots, x-family, when-not-to-use
- docs concepts (9 pages)
- docs guides (22 pages)
- docs modules (10 stable roots + 22 x/* family or subordinate primers)
- docs reference (index + 6 reference pages)
- bilingual EN/ZH parity: every marketing page and docs page ships in both locales
- locale-aware canonical and Open Graph metadata on all pages

Ongoing polish items (post-v1.1.0):

- keep `docs/reference/deprecation.mdx` and `docs/reference/stability.mdx` aligned with `docs/reference/deprecation.md` and `specs/extension-maturity.yaml` on every promotion
- track translation lag with `scripts/check-translation-lag.mjs`; the tracker only detects English-side drift, so periodic full-content passes are still needed for parity
- keep the "specific projects" section on `/use-cases` in sync with `use-cases/*` folder contents
