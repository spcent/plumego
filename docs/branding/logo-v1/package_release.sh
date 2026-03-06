#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RELEASE="$ROOT/release"

mkdir -p "$RELEASE"/{svg,png,favicon}

cp "$ROOT/final/svg/plumego-primary-horizontal-light.svg" "$RELEASE/svg/"
cp "$ROOT/final/svg/plumego-primary-horizontal-dark.svg" "$RELEASE/svg/"
cp "$ROOT/final/svg/plumego-primary-square-light.svg" "$RELEASE/svg/"
cp "$ROOT/final/svg/plumego-primary-square-dark.svg" "$RELEASE/svg/"
cp "$ROOT/final/svg/plumego-backup-horizontal-light.svg" "$RELEASE/svg/"
cp "$ROOT/final/svg/plumego-backup-horizontal-dark.svg" "$RELEASE/svg/"
cp "$ROOT/final/svg/plumego-backup-square-light.svg" "$RELEASE/svg/"
cp "$ROOT/final/svg/plumego-backup-square-dark.svg" "$RELEASE/svg/"

cp "$ROOT/final/png/plumego-primary-horizontal-light-1024.png" "$RELEASE/png/"
cp "$ROOT/final/png/plumego-primary-horizontal-light-512.png" "$RELEASE/png/"
cp "$ROOT/final/png/plumego-primary-square-light-1024.png" "$RELEASE/png/"
cp "$ROOT/final/png/plumego-primary-square-light-512.png" "$RELEASE/png/"
cp "$ROOT/final/png/plumego-backup-square-light-1024.png" "$RELEASE/png/"
cp "$ROOT/final/png/plumego-backup-square-light-512.png" "$RELEASE/png/"

cp "$ROOT/final/favicon/favicon.ico" "$RELEASE/favicon/"
cp "$ROOT/final/favicon/favicon-16.png" "$RELEASE/favicon/"
cp "$ROOT/final/favicon/favicon-32.png" "$RELEASE/favicon/"
cp "$ROOT/final/favicon/favicon-180.png" "$RELEASE/favicon/"

cat > "$RELEASE/MANIFEST.md" <<MANIFEST
# plumego logo release package

## Includes

- Primary logo: horizontal + square (light/dark), SVG
- Backup logo: horizontal + square (light/dark), SVG
- Core PNG exports for primary/backup at 1024 and 512
- Favicon set: .ico, 16, 32, 180

## Source

Generated from \

a docs/branding/logo-v1/final/*.\

Regenerate with:

\`\`\`bash
./docs/branding/logo-v1/build_assets.sh
python3 ./docs/branding/logo-v1/generate_favicon.py
./docs/branding/logo-v1/package_release.sh
\`\`\`
MANIFEST

echo "Release package created at: $RELEASE"
