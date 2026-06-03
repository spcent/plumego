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
RELEASE_BASE_URL="${RELEASE_BASE_URL:-}"

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
WAILS_BIN="$(command -v wails || true)"
if [ -z "$WAILS_BIN" ]; then
  GOPATH_BIN="$(go env GOPATH)/bin/wails"
  if [ -x "$GOPATH_BIN" ]; then
    WAILS_BIN="$GOPATH_BIN"
  fi
fi

if [ -n "$WAILS_BIN" ]; then
  echo "Using Wails CLI: $WAILS_BIN"
  echo "Building desktop app..."
  mkdir -p build
  cp internal/desktop/resources/appicon.png build/appicon.png
  rm -rf build/bin/cloud-vault-desktop build/bin/cloud-vault-desktop.app build/bin/cloud-vault.app
  if "$WAILS_BIN" build -skipbindings -tags wails -o cloud-vault-desktop; then
    desktop_app="$(find build/bin -maxdepth 1 -type d -name '*.app' | head -n 1)"
    if [ -n "$desktop_app" ]; then
      app_name="$(basename "$desktop_app")"
      cp -R "$desktop_app" "$DIST/desktop/"
      if command -v ditto &> /dev/null; then
        (cd "$DIST/desktop" && ditto -c -k --sequesterRsrc --keepParent "$app_name" cloud-vault-desktop-darwin-arm64.zip)
      fi
    elif [ -f "build/bin/cloud-vault-desktop" ]; then
      cp "build/bin/cloud-vault-desktop" "$DIST/desktop/"
    fi
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
server_sha="$(grep './server/markdown-vault' "$DIST/checksums.txt" | awk '{print $1}' || true)"
desktop_sha="$(grep './desktop/cloud-vault-desktop-darwin-arm64.zip' "$DIST/checksums.txt" | awk '{print $1}' || true)"
if [ -z "$server_sha" ]; then
  echo "Error: server checksum missing"
  exit 1
fi

notes_url=""
download_page_url=""
server_url=""
desktop_url=""
if [ -n "$RELEASE_BASE_URL" ]; then
  notes_url="${RELEASE_BASE_URL%/}/release-notes-v${VERSION}.md"
  download_page_url="${RELEASE_BASE_URL%/}/"
  server_url="${RELEASE_BASE_URL%/}/server/markdown-vault"
  desktop_url="${RELEASE_BASE_URL%/}/desktop/cloud-vault-desktop-darwin-arm64.zip"
fi

desktop_download=""
if [ -n "$desktop_sha" ]; then
  desktop_download="$(cat <<EOF_DESKTOP
,
    "darwin_arm64_desktop": {
      "url": "${desktop_url}",
      "sha256": "${desktop_sha}"
    }
EOF_DESKTOP
)"
fi

cat > "$DIST/latest.json" <<EOF
{
  "version": "${VERSION}",
  "released_at": "${BUILD_TIME}",
  "notes_url": "${notes_url}",
  "download_page_url": "${download_page_url}",
  "critical": false,
  "downloads": {
    "linux_amd64_server": {
      "url": "${server_url}",
      "sha256": "${server_sha}"
    }${desktop_download}
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
