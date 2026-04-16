# Plumego Website

This directory contains the official Plumego website.

## Goals

- fully static output
- deployable to Cloudflare Workers Static Assets
- Markdown/MDX-first publishing workflow
- bilingual support for English and Chinese
- shared light/dark theme across marketing pages and docs

## Stack

- Astro
- Starlight
- MDX
- Cloudflare Workers Static Assets

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

- production: `plumego.dev`
- www redirect or secondary host: `www.plumego.dev`

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

## Initial Scope

Phase 1 includes:

- home page
- docs landing page
- getting started
- reference app
- modules overview
- roadmap
- releases

Out of scope for Phase 1:

- blog
- SSR search
- playground
- account features
- API explorer
