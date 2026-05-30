#!/usr/bin/env bash
# Release script for Cloud Vault v1.0
# Generates checksums.txt and latest.json for auto-update check
# Usage: ./scripts/release-v1.sh [VERSION]
# Example: ./scripts/release-v1.sh 1.0.0

set -euo pipefail

VERSION="${1:-1.0.0}"
COMMIT=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
CHANNEL="stable"
DIST="dist"

echo "Building Cloud Vault v${VERSION} (${COMMIT}) at ${BUILD_TIME}..."

# Clean
echo "Cleaning dist directory..."
rm -rf "$DIST"
mkdir -p "$DIST/server" "$DIST/desktop"

# Tests
echo "Running full test suite..."
go test ./... -count=1

# Frontend
echo "Building frontend..."
cd web
if command -v pnpm &> /dev/null; then
  pnpm install --frozen-lockfile
  pnpm build
elif command -v npm &> /dev/null; then
  npm ci
  npm run build
else
  echo "Error: Neither pnpm nor npm found. Cannot build frontend."
  exit 1
fi
cd ..

# LDFLAGS
LDFLAGS="-s -w \
  -X cloud-vault/internal/version.Version=${VERSION} \
  -X cloud-vault/internal/version.Commit=${COMMIT} \
  -X cloud-vault/internal/version.BuildTime=${BUILD_TIME} \
  -X cloud-vault/internal/version.Channel=${CHANNEL}"

# Server binary
echo "Building server binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags "$LDFLAGS" -o "$DIST/server/markdown-vault" ./cmd/server

# Server package
echo "Packaging server..."
cp config.example.toml "$DIST/server/"
cp env.example "$DIST/server/.env.example" 2>/dev/null || true
cp README.md "$DIST/server/"
[ -f LICENSE ] && cp LICENSE "$DIST/server/"
[ -f docs/release-notes.md ] && cp docs/release-notes.md "$DIST/server/"

# Desktop (if Wails available)
if command -v wails &> /dev/null; then
  echo "Building desktop app..."
  if wails build -o "$DIST/desktop/cloud-vault-desktop"; then
    echo "Desktop build successful"
  else
    echo "Warning: Wails build failed, continuing without desktop app"
  fi
else
  echo "Warning: Wails CLI not found, skipping desktop build"
  echo "Install Wails with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
fi

# Checksums
echo "Generating checksums..."
cd "$DIST"
find . -type f \( -name "markdown-vault" -o -name "*.dmg" -o -name "*.exe" -o -name "*.AppImage" -o -name "*.zip" \) \
  -exec sha256sum {} \; > checksums.txt
cd ..

# Latest.json
echo "Generating latest.json..."
cat > "$DIST/latest.json" <<EOF
{
  "version": "${VERSION}",
  "released_at": "${BUILD_TIME}",
  "notes_url": "https://example.com/releases/v${VERSION}",
  "download_page_url": "https://example.com/download",
  "critical": false,
  "downloads": {
    "darwin_arm64": {
      "url": "https://example.com/downloads/Cloud-Vault-v${VERSION}-macOS-arm64.dmg",
      "sha256": "$(grep 'darwin_arm64' "$DIST/checksums.txt" | awk '{print $1}' || echo 'placeholder')"
    },
    "darwin_amd64": {
      "url": "https://example.com/downloads/Cloud-Vault-v${VERSION}-macOS-amd64.dmg",
      "sha256": "$(grep 'darwin_amd64' "$DIST/checksums.txt" | awk '{print $1}' || echo 'placeholder')"
    },
    "windows_amd64": {
      "url": "https://example.com/downloads/Cloud-Vault-v${VERSION}-windows-amd64.exe",
      "sha256": "$(grep 'windows_amd64' "$DIST/checksums.txt" | awk '{print $1}' || echo 'placeholder')"
    },
    "linux_amd64": {
      "url": "https://example.com/downloads/Cloud-Vault-v${VERSION}-linux-amd64.AppImage",
      "sha256": "$(grep 'linux_amd64' "$DIST/checksums.txt" | awk '{print $1}' || echo 'placeholder')"
    }
  }
}
EOF

# Release notes
if [ -f docs/release-notes.md ]; then
  cp docs/release-notes.md "$DIST/release-notes-v${VERSION}.md"
  echo "Copied release notes"
fi

echo ""
echo "Release v${VERSION} complete!"
echo ""
echo "Artifacts in: $DIST/"
ls -lh "$DIST/"
echo ""
echo "Checksums:"
cat "$DIST/checksums.txt"
echo ""
echo "Latest.json:"
cat "$DIST/latest.json"
echo ""
echo "Next steps:"
echo "1. Upload artifacts to your release server"
echo "2. Update latest.json URL in config.toml"
echo "3. Tag release: git tag v${VERSION}"
echo "4. Push tag: git push origin v${VERSION}"
