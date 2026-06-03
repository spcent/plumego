#!/usr/bin/env bash
# Smoke test for release build
# Usage: ./scripts/release_test.sh [VERSION]
# Starts the built binary, hits /api/v1/health, then shuts it down.

set -euo pipefail

VERSION="${1:-dev}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RELEASE_DIR="${REPO_ROOT}/dist/cloud-vault-${VERSION}"
PACKAGED_BINARY="${REPO_ROOT}/dist/server/markdown-vault"
PORT=18081

echo "=== Release Smoke Test ==="
echo "Version: ${VERSION}"
echo "Packaged binary: ${PACKAGED_BINARY}"
echo ""

if [ ! -f "${PACKAGED_BINARY}" ] && [ ! -f "${RELEASE_DIR}/cloud-vault" ]; then
    echo "ERROR: Packaged binary not found at ${PACKAGED_BINARY} or ${RELEASE_DIR}/cloud-vault"
    echo "Run ./scripts/release-v1.sh ${VERSION} first"
    exit 1
fi

# Create a temp data directory for the test
TEST_DIR="$(mktemp -d)"
trap 'echo "Cleaning up test dir: ${TEST_DIR}"; rm -rf "${TEST_DIR}"' EXIT

BINARY="${TEST_DIR}/markdown-vault-smoke"
echo "Building host smoke binary: ${BINARY}"
(cd "${REPO_ROOT}" && go build -ldflags="-X cloud-vault/internal/version.Version=${VERSION}" -o "${BINARY}" ./cmd/server)

# Create minimal config
cat > "${TEST_DIR}/config.toml" <<EOF
[server]
addr = ":${PORT}"

[database]
path = "${TEST_DIR}/data/app.db"

[storage]
provider = "local"

[storage.local]
root = "${TEST_DIR}/data/objects"

[app]
max_upload_size_mb = 10
version_policy = "bounded"
version_keep_latest = 5

[auth]
enabled = false
EOF

cd "${TEST_DIR}"

echo "[1/3] Starting server..."
"${BINARY}" --config "${TEST_DIR}/config.toml" &
SERVER_PID=$!

# Wait for server to start
echo "[2/3] Waiting for server to start..."
for i in {1..30}; do
    if curl -s "http://localhost:${PORT}/api/v1/health" > /dev/null 2>&1; then
        echo "Server started after ${i} seconds"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "ERROR: Server did not start within 30 seconds"
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi
    sleep 1
done

# Test health endpoint
echo "[3/3] Testing /api/v1/health endpoint..."
RESPONSE=$(curl -s "http://localhost:${PORT}/api/v1/health")
echo "Response: ${RESPONSE}"

if echo "${RESPONSE}" | grep -q '"status"'; then
    echo "✓ Health endpoint returned valid JSON"
else
    echo "✗ Health endpoint response unexpected"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

# Verify version matches
if echo "${RESPONSE}" | grep -q "${VERSION}"; then
    echo "✓ Version matches: ${VERSION}"
else
    echo "⚠ Version mismatch in response (expected ${VERSION})"
fi

# Shutdown
echo ""
echo "Shutting down server..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo ""
echo "=== Smoke Test Passed ==="
