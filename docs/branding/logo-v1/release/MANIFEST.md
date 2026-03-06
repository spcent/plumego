# plumego logo release package

## Includes

- Primary logo: horizontal + square (light/dark), SVG
- Backup logo: horizontal + square (light/dark), SVG
- Core PNG exports for primary/backup at 1024 and 512
- Favicon set: .ico, 16, 32, 180

## Source

Generated from `docs/branding/logo-v1/final/*`.

Regenerate with:

```bash
./docs/branding/logo-v1/build_assets.sh
python3 ./docs/branding/logo-v1/generate_favicon.py
./docs/branding/logo-v1/package_release.sh
```
