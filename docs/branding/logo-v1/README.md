# plumego logo-v1 assets

This directory contains the full logo design package for plumego v1.

## Deliverables

- 3 directions: `monogram`, `toolkit`, `wordmark`
- 2 layouts per direction: `square`, `horizontal`
- 3 variants per layout: `color`, `bw`, `reverse`
- Source vectors: SVG under each `*/svg/`
- Rasters: PNG sizes `1024`, `512`, `256`, `128` under each `*/png/`
- Final selected set: `final/svg/` and `final/png/`
- Favicon package: `final/favicon/` (`favicon.ico`, png 16/32/48/64/180/192/512)
- Release-ready slim package: `release/` (primary+backup essentials only)

## Selection result

- Primary logo: `wordmark` (Wordmark First)
- Backup logo: `monogram` (P + Plum)

## Build

```bash
./docs/branding/logo-v1/build_assets.sh
python3 ./docs/branding/logo-v1/generate_favicon.py
./docs/branding/logo-v1/package_release.sh
```
