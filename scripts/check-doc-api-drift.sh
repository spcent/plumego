#!/usr/bin/env bash

set -euo pipefail

stale='(core|plumego)\.WithRecommendedMiddleware\(|(core|plumego)\.WithTenantMiddleware\(|(core|plumego)\.WithTenantConfigManager\(|core\.WithRecovery\(|core\.WithLogging\(|app\.Group\(|core\.WithServer\('

if [ "$#" -gt 0 ]; then
  targets=("$@")
else
  targets=(
    README.md
    README_CN.md
    docs/README.md
    docs/getting-started.md
    docs/modules/core/*.md
    docs/modules/router/*.md
    docs/modules/middleware/*.md
    docs/modules/ai/*.md
    docs/modules/log/*.md
    docs/modules/store/*.md
  )
fi

if rg -n -S -e "$stale" "${targets[@]}"; then
  echo "Found removed APIs in canonical docs. Migrate snippets to current v1 canonical style."
  exit 1
fi

echo "doc-api-drift: ok"
