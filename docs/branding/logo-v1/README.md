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

## Selection result

- Primary logo: `wordmark` (Wordmark First)
- Backup logo: `monogram` (P + Plum)
- Rationale:
  - Better title readability in README and document cover headers
  - Strong compatibility with technical branding and Go ecosystem visuals
  - Maintains recognition at 128px and acceptable clarity at 32px icon tests

## Build

Regenerate all assets:

```bash
./docs/branding/logo-v1/build_assets.sh
python3 ./docs/branding/logo-v1/generate_favicon.py
```

This script regenerates all SVG variants and all required PNG exports.
