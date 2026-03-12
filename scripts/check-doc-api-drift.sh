#!/usr/bin/env bash

set -euo pipefail

stale='(core|plumego)\.WithRecommendedMiddleware\(|(core|plumego)\.WithTenantMiddleware\(|(core|plumego)\.WithTenantConfigManager\(|core\.WithRecovery\(|core\.WithLogging\(|app\.Group\(|core\.WithServer\(|health\.ReadinessHandler\(\)|health\.SetReady\(|health\.NewChecker\(|health\.LivenessHandler\(|checker\.ReadinessHandler\(|URL\([^)]*map\[string\]string|[A-Za-z0-9_]+,\s*err\s*:=\s*[A-Za-z0-9_]+\.URL\('

if [ "$#" -gt 0 ]; then
  targets=("$@")
else
  targets=(
    README.md
    README_CN.md
    docs/getting-started.md
    docs/modules/core/*.md
    docs/modules/router/*.md
    docs/modules/middleware/*.md
    docs/modules/health/*.md
    docs/modules/x-ai/*.md
    docs/modules/log/*.md
    docs/modules/store/*.md
  )
  while IFS= read -r file; do
    targets+=("$file")
  done < <(find examples/docs/en examples/docs/zh -type f -name '*.md' 2>/dev/null)
fi

if rg -n -S -e "$stale" "${targets[@]}"; then
  echo "Found removed APIs in canonical docs. Migrate snippets to current v1 canonical style."
  exit 1
fi

echo "doc-api-drift: ok"
