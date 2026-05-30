#!/usr/bin/env bash
# Release build script for Cloud Vault V0.8
# Usage: ./scripts/release.sh [VERSION]
# Example: ./scripts/release.sh 0.8.0

set -euo pipefail

# Configuration
VERSION="${1:-dev}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${REPO_ROOT}/dist"
RELEASE_DIR="${DIST_DIR}/cloud-vault-${VERSION}"
ARCHIVE_NAME="cloud-vault-${VERSION}-linux-amd64.tar.gz"

echo "=== Cloud Vault Release Build ==="
echo "Version: ${VERSION}"
echo "Repository: ${REPO_ROOT}"
echo ""

# Clean previous build
echo "[1/7] Cleaning previous build..."
rm -rf "${DIST_DIR}"
mkdir -p "${RELEASE_DIR}"

# Build frontend
echo "[2/7] Building frontend..."
cd "${REPO_ROOT}/web"
if [ ! -f "package.json" ]; then
    echo "ERROR: web/package.json not found"
    exit 1
fi

# Install dependencies
if command -v pnpm &> /dev/null; then
    pnpm install --frozen-lockfile
elif command -v npm &> /dev/null; then
    npm ci
else
    echo "ERROR: Neither pnpm nor npm found"
    exit 1
fi

# Build
if command -v pnpm &> /dev/null; then
    pnpm build
else
    npm run build
fi

# Copy frontend build output to internal/web/static/
echo "[3/7] Copying frontend assets to Go embed directory..."
STATIC_DIR="${REPO_ROOT}/internal/web/static"
rm -rf "${STATIC_DIR}"
mkdir -p "${STATIC_DIR}"
cp -r "${REPO_ROOT}/web/dist/"* "${STATIC_DIR}/"

# Run Go tests
echo "[4/7] Running Go tests..."
cd "${REPO_ROOT}"
go test ./... -count=1

# Build Go binary
echo "[5/7] Building Go binary..."
cd "${REPO_ROOT}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags "-s -w -X main.version=${VERSION}" \
  -o "${RELEASE_DIR}/cloud-vault" ./cmd/server

# Copy configuration files and documentation
echo "[6/7] Copying configuration files..."
cp "${REPO_ROOT}/config.example.toml" "${RELEASE_DIR}/"
cp "${REPO_ROOT}/env.example" "${RELEASE_DIR}/.env.example"
cp "${REPO_ROOT}/README.md" "${RELEASE_DIR}/"
[ -f "${REPO_ROOT}/LICENSE" ] && cp "${REPO_ROOT}/LICENSE" "${RELEASE_DIR}/"

# Create archive
echo "[7/7] Creating release archive..."
cd "${DIST_DIR}"
tar -czf "${ARCHIVE_NAME}" "cloud-vault-${VERSION}"

echo ""
echo "=== Release Build Complete ==="
echo "Archive: ${DIST_DIR}/${ARCHIVE_NAME}"
echo "Binary: ${RELEASE_DIR}/cloud-vault"
echo ""
echo "Archive contents:"
tar -tzf "${ARCHIVE_NAME}" | head -20
echo ""
echo "To deploy:"
echo "  1. Extract: tar -xzf ${ARCHIVE_NAME}"
echo "  2. Copy config.example.toml to config.toml and edit"
echo "  3. Copy .env.example to .env and edit"
echo "  4. Run: ./cloud-vault"
